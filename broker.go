package eventlogger

import "context"

// Broker
type Broker struct {
	Graphs map[EventType][]*Graph
}

func (b *Broker) Process(ctx context.Context, t EventType, payload PayloadType) error {
	return nil
}
