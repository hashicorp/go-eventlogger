package eventlogger

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"
)

// Gateable defines an interface for Event payloads which are "gateable" by
// the GatedFilter
type Gateable interface {
	// GetID returns an ID which allows the GatedFilter to determine that the
	// payload is part of a group of Gateable payloads.
	GetID() string

	// FlushEvent returns true when the Gateable event payload includes a Flush
	// indicator.
	FlushEvent() bool

	// ComposeFrom creates one payload which is a composition of a list events.
	// When ComposeFrom(...) is called by a GatedFilter the receiver will
	// always be nil. The payload returned must not have a Gateable payload.
	ComposeFrom(events []*Event) (t EventType, payload interface{}, err error)
}

// gatedEvent is a list of Events with the same Gateable.GetID().  These events
// have an exp (expiration) time. A gatedEvent has a list.Element, which allows
// it to be part of a linked list of gatedEvents.
type gatedEvent struct {
	// id of the event and all the "events"
	id string
	// events are an ordered list that all have the same ID
	events []*Event
	// exp of the gatedEvent
	exp time.Time
	// element of a linked list of gatedEvents
	element *list.Element
}

// DefaultGatedEventTimeout defines a default expiry for events processed by a
// GatedFilter
const DefaultGatedEventTimeout = time.Second * 10

// GatedFilter provides the ability to buffer events identified by a
// Gateable.GetID() until an event is processed that returns true for
// Gateable.FlushEvent().
//
// When a Gateable Event returns true for FlushEvent(), the filter will call
// Gateable.ComposedOf(...) with the list of gated events with the coresponding
// Gateable.GetID() up to that point in time and return the resulting composed
// event.   There is no dependency on GatedFilter.Broker to handle an event that
// returns true for FlushEvent() since the GatedFilter simply needs to return
// the flushed event from GatedFilter.Process(...)

// GatedFilter.Broker is only used when handling expired events or when
// handling calls to GatedFilter.FlushAll().  If GatedFilter.Broker is nil,
// expired gated events will simply be deleted. If the Broker is NOT nil, then
// the expiring gated events will be flushed using Gateable.ComposedOf(...) and
// the resulting composed event is sent using the Broker.  If the Broker is nil
// when GatedFilter.FlushAll() is called then the gated events will just be
// deleted.  If the Broker is not nil when GatedFilter.FlushAll() is called,
// then all the gated events will be sent using the Broker.
type GatedFilter struct {
	// Broker used to send along expired gated events
	Broker *Broker

	// Expiration for gated events.  It's important because without an
	// expiration gated events that aren't flushed/processed could consume all
	// available memory.  Expired events will be sent along if there's a Broker
	// or deleted if there's no Broker. If no expiration is set the
	// DefaultGatedEventTimeout will be used.
	Expiration time.Duration

	// NowFunc is a func that returns the current time and the GatedFilter and
	// if unset, it will default to time.Now()
	NowFunc func() time.Time

	// gated uses Gateable.GetID() to uniquely identify gatedEvents (collections of Gatable
	// payloads)
	gated map[string]*gatedEvent

	// orderedGated gives us an ordered (by timestamp) linked list of gated
	// events, so we can efficiently process expired entries.
	orderedGated *list.List

	// composedFrom is a reference to the Gateable.ComposedFrom func for
	// the specific type of Gateable event
	composeFrom func(events []*Event) (t EventType, payload interface{}, e error)
	l           sync.RWMutex
}

var _ Node = &GatedFilter{}

// Process will determine if an Event is Gateable.  Events that are not not
// Gateable are immediately returned. If the Event is Gateable, it's added to a
// list of Events using it's Gateable.ID() as an index, until an event with a
// matching Gateable.ID() is processed where Gateable.Flush() returns true.  If
// Gateable.Flush(), then Gateable.ComposedFrom([]*Event) is called with all the
// gated events for the ID.
func (w *GatedFilter) Process(ctx context.Context, e *Event) (*Event, error) {
	const op = "eventlogger.(GatedWriter).Process"
	if e == nil {
		return nil, fmt.Errorf("%s: missing event: %w", op, ErrInvalidParameter)
	}
	g, ok := e.Payload.(Gateable)
	if !ok {
		// the event isn't gateable so just let it proceed along its merry way
		// in the pipeline
		return e, nil
	}
	if g.GetID() == "" {
		return nil, fmt.Errorf("%s: missing ID: %w", op, ErrInvalidParameter)
	}
	w.l.Lock()
	// since there's no factory, we need to make sure the GatedFilter is
	// initialized properly
	if w.gated == nil {
		w.gated = map[string]*gatedEvent{}
	}
	if w.orderedGated == nil {
		w.orderedGated = list.New()
	}
	if w.Expiration == 0 {
		w.Expiration = DefaultGatedEventTimeout
	}
	if w.composeFrom == nil {
		w.composeFrom = g.ComposeFrom
	}
	w.l.Unlock()

	// before we do much of anything else, let's take care of any expiring Gated
	// events.  Note: processExpiredEvents will acquire a lock, so we must
	// unsure the GatedFilter is unlocked before calling the func.
	if err := w.processExpiredEvents(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	w.l.Lock()
	defer w.l.Unlock()
	// Is it first time we've seen this gated event ID?
	if _, ok := w.gated[g.GetID()]; !ok {
		ge := &gatedEvent{
			id:     g.GetID(),
			events: []*Event{},
			exp:    w.Now().Add(w.Expiration),
		}
		ge.element = w.orderedGated.PushBack(ge)
		w.gated[g.GetID()] = ge
	}
	// append the inbound event to our existing events for this ID
	w.gated[g.GetID()].events = append(w.gated[g.GetID()].events, e)

	// Is this event a signal to FlushEvent?
	if g.FlushEvent() {
		// need to remove this ID, even if there's an error during composition.
		defer w.orderedGated.Remove(w.gated[g.GetID()].element)
		defer delete(w.gated, g.GetID())

		t, p, err := w.composeFrom(w.gated[g.GetID()].events)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return &Event{
			Type:      t,
			Payload:   p,
			CreatedAt: w.Now(),
			Formatted: make(map[string][]byte),
		}, nil
	}

	return nil, nil
}

// processExpiredEvents will check gated events for expiry and send them along
// to the Broker as they expire.  If the GatedFilter has no broker, the expired
// events are just deleted.
func (w *GatedFilter) processExpiredEvents(ctx context.Context) error {
	const op = "eventlogger.(GatedFilter).ProcessExpiredEvents"
	w.l.Lock()
	defer w.l.Unlock()
	if w.composeFrom == nil {
		return fmt.Errorf("%s: composedFrom func is not initialized: %w", op, ErrInvalidParameter)
	}
	if w.orderedGated == nil {
		return nil
	}
	if len(w.gated) == 0 {
		return nil
	}
	if w.Expiration == 0 {
		w.Expiration = DefaultGatedEventTimeout
	}

	// Iterate through list, starting with the oldest gated event at the front.
	for e := w.orderedGated.Front(); e != nil; e = e.Next() {
		ge := e.Value.(*gatedEvent)
		switch {
		case w.Now().After(ge.exp):
			if err := w.openGate(ctx, ge); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		default:
			// since the event are ordered by when they arrived, once we hit one
			// that's not expired we're done.
			break
		}
	}
	return nil
}

// FlushAll will flush all events that have been gated and is useful for
// circumstances where the system is shuting down and you need to flush
// everything that's been gated.
//
// If the Broker is nil when GatedFilter.FlushAll() is called then the gated
// events will just be deleted.  If the Broker is not nil when
// GatedFilter.FlushAll() is called, then all the gated events will be sent
// using the Broker.
func (w *GatedFilter) FlushAll(ctx context.Context) error {
	const op = "eventlogger.(GatedFilter).FlushAll"
	w.l.Lock()
	defer w.l.Unlock()
	if len(w.gated) == 0 {
		return nil
	}
	if w.composeFrom == nil {
		return fmt.Errorf("%s: composedFrom func is not initialized: %w", op, ErrInvalidParameter)
	}

	if w.Broker == nil {
		// no op... perhaps we should log this somehow in the future if the
		// GatedFilter adds a logger.  For now, we'll just drop all the events
		// into the bit bucket to nowhere.
		w.gated = nil
		w.orderedGated = nil
	}

	// Iterate through list, starting with the oldest gated event at the front.
	for e := w.orderedGated.Front(); e != nil; e = e.Next() {
		ge := e.Value.(*gatedEvent)
		if err := w.openGate(ctx, ge); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

// openGate will not acquire it's own lock, so the caller must do so before
// calling it.
func (w *GatedFilter) openGate(ctx context.Context, ge *gatedEvent) error {
	const op = "eventlogger.(GatedFilter).openGate"
	if ge == nil {
		return fmt.Errorf("%s: missing gated event: %w", op, ErrInvalidParameter)
	}
	if w.composeFrom == nil {
		return fmt.Errorf("%s: composedFrom func is not initialized: %w", op, ErrInvalidParameter)
	}
	// need to remove this, even if there's an error during composition
	defer w.orderedGated.Remove(ge.element)
	defer delete(w.gated, ge.element.Value.(*gatedEvent).id)

	t, p, err := w.composeFrom(ge.events)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, ok := p.(Gateable)
	if ok {
		// a Gateable payload would create infinite loop.
		return fmt.Errorf("%s: %T.ComposeFrom returned a Gateable payload", op, p)
	}
	switch {
	case w.Broker == nil:
		// no op... perhaps we should log this somehow in the future if
		// the GatedFilter adds a logger.  For now, we'll just drop the
		// event into the bit bucket to nowhere.
	default:
		if _, err := w.Broker.Send(ctx, t, p); err != nil {
			return err
		}
	}
	return nil
}

// Reopen is a no op for GatedFilters.
func (w *GatedFilter) Reopen() error {
	return nil
}

// Type describes the type of the node as a Filter.
func (w *GatedFilter) Type() NodeType {
	return NodeTypeFilter
}

// Now returns the current time.  If GatedFilter.NowFunc is unset, then
// time.Now() is used as a default.
func (w *GatedFilter) Now() time.Time {
	if w.NowFunc != nil {
		return w.NowFunc()
	}
	return time.Now()
}

// SimpleGatedPayload implements the Gateable interface for an Event payload and
// can be used when sending events with a Broker.
type SimpleGatedPayload struct {
	// ID must be a unique ID
	ID string `json:"id"`

	// Flush value is returned from FlushEvent()
	Flush bool `json:"-"`

	// Header is top level header info
	Header map[string]interface{} `json:"header,omitempty"`

	// Detail is detail info
	Detail map[string]interface{} `json:"detail,omitempty"`
}

var _ Gateable = &SimpleGatedPayload{}

// GetID returns the unique ID
func (s *SimpleGatedPayload) GetID() string {
	return s.ID
}

// FlushEvent tells the GatedFilter to flush/process the events associated with
// the Gateable ID
func (s *SimpleGatedPayload) FlushEvent() bool {
	return s.Flush
}

// SimpleGatedEventDetails defines the struct used in the
// SimpleGatedEventPayload.Details list.
type SimpleGatedEventDetails struct {
	Type      string                 `json:"type"`
	CreatedAt string                 `json:"created_at"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// SimpleGatedEventPayload defines the resulting Event from SimpleGatedPayload.ComposeFrom
type SimpleGatedEventPayload struct {
	ID      string                    `json:"id"`
	Header  map[string]interface{}    `json:"header,omitempty"`
	Details []SimpleGatedEventDetails `json:"details,omitempty"`
}

// ComposedFrom will build a single event payload which will be
// Flushed/Processed from a collection of gated events.  The payload returned is
// not a Gateable payload intentionally.  Note: the SimpleGatedPayload receiver
// is always nil when this function is called.
func (s *SimpleGatedPayload) ComposeFrom(events []*Event) (EventType, interface{}, error) {
	const op = "eventlogger.(SimpleGatedPayload).ComposedFrom"
	if len(events) == 0 {
		return "", nil, fmt.Errorf("%s: missing events: %w", op, ErrInvalidParameter)
	}

	payload := SimpleGatedEventPayload{}
	for i, v := range events {
		g, ok := v.Payload.(*SimpleGatedPayload)
		if !ok {
			return "", nil, fmt.Errorf("%s: event %d is not a simple gated payload: %w", op, i, ErrInvalidParameter)
		}
		payload.ID = g.GetID()
		if g.Header != nil {
			for hdrK, hdrV := range g.Header {
				if payload.Header == nil {
					payload.Header = map[string]interface{}{}
				}
				payload.Header[hdrK] = hdrV
			}
		}
		if g.Detail != nil {
			payload.Details = append(payload.Details, SimpleGatedEventDetails{
				Type:      string(v.Type),
				CreatedAt: v.CreatedAt.String(),
				Payload:   g.Detail,
			})
		}
	}
	return events[0].Type, payload, nil
}
