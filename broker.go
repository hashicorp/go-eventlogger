package eventlogger

// Future.  This should probably be a channel or some such.
type Future interface {
	Await() error
}

// Broker
type Broker struct {
	Graphs map[EventType]*Graph
}

func (b *Broker) Process(e *Event) Future {
	return nil
}
