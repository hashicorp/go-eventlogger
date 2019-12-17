package eventlogger

// Graph
type Graph struct {
	Root Node
	// GuaranteeLevel -- future enhancement
}

// Future
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
