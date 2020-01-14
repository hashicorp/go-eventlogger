package eventlogger

import (
	"context"
)

// Predicate returns true if we want to keep the Event.
type Predicate func(e *Event) (bool, error)

// Filter
type Filter struct {
	nodes     []Node
	Predicate Predicate
	name      string
}

var _ LinkableNode = &Filter{}

func (f *Filter) Process(ctx context.Context, e *Event) (*Event, error) {

	// Use the predicate to see if we want to keep the event.
	keep, err := f.Predicate(e)
	if err != nil {
		return nil, err
	}
	if !keep {
		// Return nil to signal that the event should be discarded.
		return nil, nil
	}

	// return the event
	return e, nil
}

func (f *Filter) SetNext(nodes []Node) {
	f.nodes = nodes
}

func (f *Filter) Next() []Node {
	return f.nodes
}

func (f *Filter) Reopen() error {
	return nil
}

func (f *Filter) Type() NodeType {
	return NodeTypeFilter
}

func (f *Filter) Name() string {
	return f.name
}
