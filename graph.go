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
