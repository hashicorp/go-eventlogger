package eventlogger

import "context"

// Graph
type Graph struct {
	Root Node
}

// Process the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(ctx context.Context, e *Event) error {
	return g.process(ctx, g.Root, e)
}

// Recursively process every node in the graph.
func (g *Graph) process(ctx context.Context, node Node, e *Event) error {

	// Process the current Node
	e, err := node.Process(e)
	if err != nil {
		return err
	}

	// If the new Event is nil, it has been filtered out and we are done.
	if e == nil {
		return nil
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		for _, child := range ln.Next() {

			err = g.process(ctx, child, e)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
