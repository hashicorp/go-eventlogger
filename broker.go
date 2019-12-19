package eventlogger

import (
	"context"
	"fmt"
	"time"
)

// Broker
type Broker struct {
	Graphs map[EventType]*Graph
}

func (b *Broker) Send(ctx context.Context, t EventType, payload PayloadType) error {

	g, ok := b.Graphs[t]
	if !ok {
		return fmt.Errorf("No Graph for EventType %s", t)
	}

	e := &Event{
		Type:      t,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
		Payload:   payload,
	}

	return g.Process(ctx, e)
}
