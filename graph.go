package eventlogger

import (
	"context"
	"fmt"
)

// Graph
type Graph struct {
	Root Node
}

// Process the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(ctx context.Context, e *Event) (Status, error) {
	return g.process(ctx, g.Root, e)
}

// Recursively process every node in the graph.
func (g *Graph) process(ctx context.Context, node Node, e *Event) (Status, error) {

	// Process the current Node
	e, err := node.Process(e)
	if err != nil {
		return Status{Warnings: []error{err}}, err
	}

	var s Status
	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		// If the new Event is nil, it has been filtered out and we are done.
		if e == nil {
			return Status{}, nil
		}

		for _, child := range ln.Next() {
			status, _ := g.process(ctx, child, e)
			s.Warnings = append(s.Warnings, status.Warnings...)
			s.SentToSinks = append(s.SentToSinks, status.SentToSinks...)
		}
	} else {
		return Status{SentToSinks: []string{node.Name()}}, nil
	}

	if len(s.Warnings) > 0 && len(s.SentToSinks) == 0 {
		// TODO: the following would make sense:
		//  return s, fmt.Errorf("event not written to any sinks successfully")

		// For now we're not doing the above: instead, even when we failed to
		// log the event, we're returning a nil error.  Rationale: we don't yet
		// have support for configurable delivery policies, and if the user has
		// defined only a single bogus sink, and modifying the sink config fails
		// due to failure to log the reconfiguration event because of that bogus
		// sink, the user will never be able to repair the configuration.

		// We'll have delivery policies soon enough, and in the meantime callers
		// can log the warnings to inform their users of any problems logging.
	}
	return s, nil
}

func (g *Graph) Reload(ctx context.Context) error {
	return g.reload(ctx, g.Root)
}

// Recursively process every node in the graph.
func (g *Graph) reload(ctx context.Context, node Node) error {

	// Process the current Node
	err := node.Reload()
	if err != nil {
		return err
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		for _, child := range ln.Next() {

			err = g.reload(ctx, child)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Graph) Validate() error {
	return g.validate(nil, g.Root)
}

func (g *Graph) validate(parent, node Node) error {
	innerNode, isInner := node.(LinkableNode)

	switch {
	case isInner && len(innerNode.Next()) == 0:
		return fmt.Errorf("non-sink node has no children")
	case !isInner && parent == nil:
		return fmt.Errorf("sink node at root")
	case !isInner && parent.Type() != NodeTypeFormatter:
		return fmt.Errorf("sink node without preceding formatter")
	case !isInner:
		return nil
	}

	// Process any child nodes.  This is depth-first.
	for _, child := range innerNode.Next() {
		err := g.validate(node, child)
		if err != nil {
			return err
		}
	}

	return nil
}
