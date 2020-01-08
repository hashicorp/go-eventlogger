package eventlogger

import (
	"context"
	"fmt"
	"time"
)

type clock struct {
	now time.Time
}

// Broker
type Broker struct {
	Graphs map[EventType]*Graph
	*clock
}

func (c *clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

func (b *Broker) Validate() error {
	if len(b.Graphs) == 0 {
		return fmt.Errorf("no graphs in broker")
	}

	for _, g := range b.Graphs {
		if err := g.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type Status struct {
	SentToSinks []string
	Warnings    []error
}

func (b *Broker) Send(ctx context.Context, t EventType, payload interface{}) (Status, error) {
	g, ok := b.Graphs[t]
	if !ok {
		return Status{}, fmt.Errorf("No Graph for EventType %s", t)
	}

	e := &Event{
		Type:      t,
		CreatedAt: b.clock.Now(),
		Formatted: make(map[string][]byte),
		Payload:   payload,
	}

	return g.Process(ctx, e)
}

func (b *Broker) Reopen(ctx context.Context) error {
	for _, g := range b.Graphs {
		if err := g.Reopen(ctx); err != nil {
			return err
		}
	}

	return nil
}
