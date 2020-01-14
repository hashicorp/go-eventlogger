package eventlogger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Broker is the top-level entity used in the library for configuring the system
// and for sending events.
type Broker struct {
	nodes  map[NodeID]Node
	graphs map[EventType]*graph
	lock   sync.RWMutex

	*clock
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		nodes:  make(map[NodeID]Node),
		graphs: make(map[EventType]*graph),
	}
}

// clock only exists to make testing simpler.
type clock struct {
	now time.Time
}

func (c *clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

// Status describes the result of a Send.
type Status struct {
	// Complete lists the names of all sinks that successfully wrote the Event.
	Complete []string
	// Warnings lists any non-fatal errors that occurred while sending an Event.
	Warnings []error
}

func (s Status) getError(threshold int) error {
	if len(s.Complete) < threshold {
		return fmt.Errorf("event not written to enough sinks")
	}
	return nil
}

// Send writes an event of type t to all configured pipelines and reports on the
// result.  An error will only be returned if configured delivery policies could
// not be satisfied.
func (b *Broker) Send(ctx context.Context, t EventType, payload interface{}) (Status, error) {

	b.lock.RLock()
	g, ok := b.graphs[t]
	b.lock.RUnlock()

	if !ok {
		return Status{}, fmt.Errorf("No graph for EventType %s", t)
	}

	e := &Event{
		Type:      t,
		CreatedAt: b.clock.Now(),
		Formatted: make(map[string][]byte),
		Payload:   payload,
	}

	return g.process(ctx, e)
}

// Reopen asks all nodes to reopen any files they have open.  This is typically
// used as part of log rotation: after rotating, the rotator sends a signal to
// the application, which then would invoke this method.
func (b *Broker) Reopen(ctx context.Context) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, g := range b.graphs {
		if err := g.reopen(ctx); err != nil {
			return err
		}
	}

	return nil
}

// NodeID is a string that uniquely identifies a Node.
type NodeID string

func (b *Broker) RegisterNode(id NodeID, node Node) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nodes[id] = node
	return nil
}

// PipelineID is a string that uniquely identifies a Pipeline within a given EventType.
type PipelineID string

// RegisterPipeline adds a pipeline to the broker.
func (b *Broker) RegisterPipeline(t EventType, id PipelineID, nodeIDs []NodeID) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{roots: make(map[PipelineID]*linkedNode)}
		b.graphs[t] = g
	}

	nodes := make([]Node, len(nodeIDs))
	for i, n := range nodeIDs {
		node, ok := b.nodes[n]
		if !ok {
			return fmt.Errorf("nodeID %q not registered", n)
		}
		nodes[i] = node
	}
	root, err := linkNodes(nodes)
	if err != nil {
		return err
	}

	err = g.doValidate(nil, root)
	if err != nil {
		return err
	}

	g.roots[id] = root

	return nil
}

// RemovePipeline removes a pipeline from the broker.
func (b *Broker) RemovePipeline(t EventType, id PipelineID) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		return fmt.Errorf("No graph for EventType %s", t)
	}

	delete(g.roots, id)
	return nil
}

// SetSuccessThreshold sets the succes threshold per eventType.  For the
// overall processing of a given event to be considered a success, at least as
// many sinks as the threshold value must successfully process the event.
func (b *Broker) SetSuccessThreshold(t EventType, successThreshold int) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if successThreshold < 0 {
		return fmt.Errorf("successThreshold must be 0 or greater")
	}

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{roots: make(map[PipelineID]*linkedNode)}
		b.graphs[t] = g
	}

	g.successThreshold = successThreshold
	return nil
}
