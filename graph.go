package eventlogger

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// graph
type graph struct {

	// roots maps PipelineIDs to root Nodes
	roots map[PipelineID]Node

	// successThreshold specifies how many sinks must store an event for Process
	// to not return an error.
	successThreshold int
}

// Process the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *graph) process(ctx context.Context, e *Event) (Status, error) {
	statusChan := make(chan Status)
	var wg sync.WaitGroup
	go func() {
		for _, root := range g.roots {
			wg.Add(1)
			g.doProcess(ctx, root, e, statusChan, &wg)
		}
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
				status.Complete = append(status.Complete, s.Complete...)
			} else {
				done = true
			}
		}
	}
	return status, status.getError(g.successThreshold)
}

// Recursively process every node in the graph.
func (g *graph) doProcess(ctx context.Context, node Node, e *Event, statusChan chan Status, wg *sync.WaitGroup) {
	defer wg.Done()

	// Process the current Node
	e, err := node.Process(e)
	if err != nil {
		select {
		case <-ctx.Done():
		case statusChan <- Status{Warnings: []error{err}}:
		}
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
			go g.doProcess(ctx, child, e, statusChan, wg)
		}
	} else {
		select {
		case <-ctx.Done():
		case statusChan <- Status{Complete: []string{node.Name()}}:
		}
	}
}

func (g *graph) reopen(ctx context.Context) error {
	var errors *multierror.Error

	for _, root := range g.roots {
		err := g.doReopen(ctx, root)
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	return errors.ErrorOrNil()
}

// Recursively reopen every node in the graph.
func (g *graph) doReopen(ctx context.Context, node Node) error {

	// Process the current Node
	err := node.Reopen()
	if err != nil {
		return err
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		for _, child := range ln.Next() {

			err = g.doReopen(ctx, child)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *graph) validate() error {
	var errors *multierror.Error

	for _, root := range g.roots {
		err := g.doValidate(nil, root)
		if err != nil {
			errors = multierror.Append(errors, err)
		}
	}

	return errors.ErrorOrNil()
}

func (g *graph) doValidate(parent, node Node) error {
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
		err := g.doValidate(node, child)
		if err != nil {
			return err
		}
	}

	return nil
}
