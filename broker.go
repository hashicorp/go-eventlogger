package eventlogger

type Graph struct {
	Root Node
	// GuaranteeLevel -- future enhancement
}

type Future interface {
	Await() error
}

type Broker struct {
	Graphs map[EventType]Graph
}

func (b *Broker) Process(e *Event) Future {
	return nil
}
