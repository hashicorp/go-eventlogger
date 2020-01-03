package eventlogger

import (
	"context"
	"fmt"
	"sync"
)

// Graph
type Graph struct {
	Root Node
	// SuccessThreshold specifies how many sinks must store an event for Process
	// to not return an error.
	SuccessThreshold int
}

func (s Status) GetError(threshold int) error {
	if len(s.SentToSinks) < threshold {
		return fmt.Errorf("event not written to enough sinks")
	}
	return nil
}

// Process the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(ctx context.Context, e *Event) (Status, error) {
	statusChan := make(chan Status)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		g.process(ctx, g.Root, e, statusChan, &wg)
		wg.Wait()
		close(statusChan)
	}()
	var status Status
	var done bool
	for !done {
		select {
		case <-ctx.Done():
			done = true
		case s, ok := <-statusChan:
			if ok {
				status.Warnings = append(status.Warnings, s.Warnings...)
				status.SentToSinks = append(status.SentToSinks, s.SentToSinks...)
			} else {
				done = true
			}
		}
	}
	return status, status.GetError(g.SuccessThreshold)
}

// Recursively process every node in the graph.
func (g *Graph) process(ctx context.Context, node Node, e *Event, statusChan chan Status, wg *sync.WaitGroup) {
	defer wg.Done()

	// Process the current Node
	e, err := node.Process(e)
	if ctx.Err() != nil {
		return
	}
	if err != nil {
		statusChan <- Status{Warnings: []error{err}}
		return
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		// If the new Event is nil, it has been filtered out and we are done.
		if e == nil {
			statusChan <- Status{}
			return
		}

		for _, child := range ln.Next() {
			wg.Add(1)
			go g.process(ctx, child, e, statusChan, wg)
		}
	} else {
		statusChan <- Status{SentToSinks: []string{node.Name()}}
	}
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
