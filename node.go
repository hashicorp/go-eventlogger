package eventlogger

// Node
type Node interface {
	Process(g *Graph, e *Event) error
}

//----------------------------------------------------------
// Filter

// Predicate returns true if we want to keep the Event.
type Predicate func(e *Event) (bool, error)

// Filter
type Filter struct {
	Predicate Predicate
	Next      Node
}

func (f *Filter) Process(g *Graph, e *Event) error {

	// Use the predicate to see if we want to keep the event.
	keep, err := f.Predicate(e)
	if err != nil {
		return err
	}
	if !keep {
		return nil
	}

	// Process the next Node
	if f.Next == nil {
		return nil
	}
	return f.Next.Process(g, e)
}

//----------------------------------------------------------
// ByteWriter

// ByteMarshaller turns an Event into a slice of bytes suitable for being
// persisted.
type ByteMarshaller func(e *Event) ([]byte, error)

// ByteWriter
type ByteWriter struct {
	Marshaller ByteMarshaller
	Next       Node
}

func (w *ByteWriter) Process(g *Graph, e *Event) error {

	// Marshal
	bytes, err := w.Marshaller(e)
	if err != nil {
		return err
	}

	// Clone the event and add in the writable representation
	ne := e.Clone()
	ne.Writable = bytes

	// Process the next Node
	if w.Next == nil {
		return nil
	}
	return w.Next.Process(g, ne)
}

////----------------------------------------------------------
////// FanOutNode creates a tree, which will be
////// useful for e.g. fanning out to multiple Sinks.
////struct FanOutNode {
////	next []Nodes
////}
