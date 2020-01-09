package eventlogger

import (
	"context"
	"fmt"
	"time"
)

// Broker
type Broker struct {
	graphs map[EventType]*graph
	*clock
}

func NewBroker() *Broker {
	return &Broker{
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

func (b *Broker) Validate() error {
	if len(b.graphs) == 0 {
		return fmt.Errorf("no graphs in broker")
	}

	for _, g := range b.graphs {
		if err := g.validate(); err != nil {
			return err
		}
	}

	return nil
}

type Status struct {
	SentToSinks []string
	Warnings    []error
}

func (s Status) GetError(threshold int) error {
	if len(s.SentToSinks) < threshold {
		return fmt.Errorf("event not written to enough sinks")
	}
	return nil
}

func (b *Broker) Send(ctx context.Context, t EventType, payload interface{}) (Status, error) {
	g, ok := b.graphs[t]
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

func (b *Broker) Reopen(ctx context.Context) error {
	for _, g := range b.graphs {
		if err := g.reopen(ctx); err != nil {
			return err
		}
	}

	return nil
}

// PipelineID is a string that uniquely identifies a Pipeline within a given EventType.
type PipelineID string

// RegisterPipeline adds a pipeline to the broker.
func (b *Broker) RegisterPipeline(t EventType, id PipelineID, root Node) error {

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{roots: make(map[PipelineID]Node)}
		b.graphs[t] = g
	}

	_, ok = g.roots[id]
	if ok {
		return fmt.Errorf("pipeline for PipelineID %s already exists", id)
	}

	err := g.doValidate(nil, root)
	if err != nil {
		return err
	}

	g.roots[id] = root

	return nil
}

// RemovePipeline removes a pipeline from the broker.
func (b *Broker) RemovePipeline(t EventType, id PipelineID) error {
	g, ok := b.graphs[t]
	if !ok {
		return fmt.Errorf("No graph for EventType %s", t)
	}

	_, ok = g.roots[id]
	if !ok {
		return fmt.Errorf("No pipeline for PipelineID %s", id)
	}

	delete(g.roots, id)
	return nil
}
