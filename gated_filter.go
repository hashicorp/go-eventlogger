package eventlogger

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"
)

// Gateable defines an interface for Event payloads which are gateable by
// the GatedFilter
type Gateable interface {
	// GetID returns an ID which allows the GatedFilter to determine that the
	// payload is part of a group of Gateable payloads.
	GetID() string

	// FlushEvents returns true when the Gateable event payload includes a Flush
	// indicator.
	FlushEvents() bool

	// ComposedFrom creates one event which is a composition of the list events
	// parameter.  The Event returned must not have a Gateable payload.
	ComposedFrom(now time.Time, events []*Event) (*Event, error)
}

type gatedEvent struct {
	id      string
	events  []*Event
	exp     time.Time
	element *list.Element
}

// DefaultGatedEventTimeout defines a default expiry for events processed by a
// GatedFilter
const DefaultGatedEventTimeout = time.Second * 10

// GatedFilter provides the ability to buffer events identified by
// Gateable.GetID() until an event is processed that returns true for
// Gateable.FlushEvents().
//
// When a Gateable Event returns true for FlushEvents(), the filter will call
// Gateable.ComposedOf(...) with the list of gated events with the
// Gateable.GetID() up to that point in time and return the resulting composed
// event.

// If GatedFilter.Broker is nill, expired gated events will simply be being
// deleted.  If the Broker is NOT nil, then the expiring gated events will be
// flushed using Gateable.ComposedOf(...) and the resulting composed event sent
// using the Broker.
type GatedFilter struct {
	// Broker used to send along expired gated events
	Broker *Broker

	// Expiration for gated events.  It's important because without an
	// expiration gated events that aren't flushed/processed could consume all
	// available memory.  Expired events will be sent along if there's a Broker
	// or deleted if there's no Broker. If no expiration is set the
	// DefaultGatedEventTimeout will be used.
	Expiration time.Duration

	// NowFunc is a time func that returns the current time and the GatedFilter
	// will default to time.Now() if it's unset.
	NowFunc func() time.Time

	l sync.RWMutex

	// gated uses Gateable.GetID() to uniquely identify collections of Gatable
	// payloads within a gatedEvent
	gated map[string]*gatedEvent

	// orderedGated gives us an ordered list of gated events, so we can efficiently process expired entries.
	orderedGated *list.List
}

var _ Node = &GatedFilter{}

// Process will call determine if the Event is Gateable.  If it's not Gateable
// then it's returned.  If the Event is Gateable, it's added to a list of Events
// for the Gateable.ID() until an event is processed where Gateable.Flush()
// returns true.  If Gateable.Flush(), then Gateable.ComposedFrom([]*Event) is
// called with all the gated events for the ID.
func (w *GatedFilter) Process(ctx context.Context, e *Event) (*Event, error) {
	const op = "eventlogger.(GatedWriter).Process"
	if e == nil {
		return nil, fmt.Errorf("%s: missing event", op)
	}
	g, ok := e.Payload.(Gateable)
	if !ok {
		// the event isn't gateable so just let it proceed along its merry way
		// in the pipeline
		return e, nil
	}

	if g.GetID() == "" {
		return nil, fmt.Errorf("%s: %s", op, "event missing ID")
	}

	w.l.Lock()
	defer w.l.Unlock()

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

	// before we do much of anything else, let's take care of any expiring Gated
	// events.
	if err := w.ProcessExpiredEvents(ctx); err != nil {
		return nil, err
	}

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

	// Is this event a signal to FlushEvents?
	if g.FlushEvents() {
		// need to remove this ID, even if there's an error during composition.
		defer w.orderedGated.Remove(w.gated[g.GetID()].element)
		defer delete(w.gated, g.GetID())

		return g.ComposedFrom(w.Now(), w.gated[g.GetID()].events)
	}

	return nil, nil
}

// ProcessExpiredEvents will check gated events for expiry and send them along
// to the Broker as they expire.  If the GatedFilter has no broker, the expired
// events are just deleted.
func (w *GatedFilter) ProcessExpiredEvents(ctx context.Context) error {
	const op = "eventlogger.(GatedFilter).ProcessExpiredEvents"
	if w.orderedGated == nil {
		return nil
	}
	if w.Expiration == 0 {
		w.Expiration = DefaultGatedEventTimeout
	}
	// Iterate through list, starting with the oldest gated event at the front.
	for e := w.orderedGated.Front(); e != nil; e = e.Next() {
		ge := e.Value.(*gatedEvent)
		if w.Now().After(ge.exp) {
			// need to remove this, even if there's an error during composition
			defer w.orderedGated.Remove(ge.element)
			defer delete(w.gated, ge.element.Value.(*gatedEvent).id)

			// well, it's one way to have a static method in Go...
			tmp := &SimpleGatedPayload{}
			e, err := tmp.ComposedFrom(w.Now(), ge.events)
			if err != nil {
				return err
			}
			switch {
			case w.Broker == nil:
				// no op... perhaps we should log this somehow in the future if
				// the GatedFilter adds a logger.  For now, we'll just drop the
				// event into the bit bucket to nowhere.
			default:
				if _, err := w.Broker.SendEvent(ctx, e); err != nil {
					return err
				}
			}
		} else {
			// since the event are ordered by when they arrived, once we hit one
			// that's not expired we're done.
			break
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

// SimpleGatedPayload defines a Gateable payload implementation.
type SimpleGatedPayload struct {
	// ID must be a unique ID
	ID string

	// Flush value is returned from FlushEvents()
	Flush bool

	// Header is top level header info
	Header map[string]interface{}

	// Detail is detail info
	Detail map[string]interface{}
}

var _ Gateable = &SimpleGatedPayload{}

// GetID returns the unique ID
func (s *SimpleGatedPayload) GetID() string {
	return s.ID
}

// FlushEvents tells the GatedFilter to flush/process the events associated with
// the Gateable ID
func (s *SimpleGatedPayload) FlushEvents() bool {
	return s.Flush
}

// ComposedFrom will build a single event which will be Flushed/Processed from a
// collection of gated events.  The event returned does not contain a Gateable
// payload intentionally.
func (s *SimpleGatedPayload) ComposedFrom(now time.Time, events []*Event) (*Event, error) {
	const op = "eventlogger.(SimpleGatedPayload).ComposedFrom"
	if now.IsZero() {
		return nil, fmt.Errorf("%s: missing now", op)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("%s: missing events", op)
	}
	payload := struct {
		ID      string
		Header  map[string]interface{}
		Details []*Event
	}{}
	for i, v := range events {
		g, ok := v.Payload.(*SimpleGatedPayload)
		if !ok {
			return nil, fmt.Errorf("%s: event %d is not a simple gated payload", op, i)
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
			payload.Details = append(payload.Details, &Event{
				Type:      v.Type,
				CreatedAt: v.CreatedAt,
				Payload:   g.Detail,
			})
		}
	}
	return &Event{
		Type:      events[0].Type,
		CreatedAt: now,
		Payload:   payload,
	}, nil
}
