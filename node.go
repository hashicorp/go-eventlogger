package eventlogger

import (
	"context"
	"errors"
	"fmt"
)

type NodeType int

const (
	_ NodeType = iota
	NodeTypeFilter
	NodeTypeFormatter
	NodeTypeSink
)

// A Node in a graph
type Node interface {
	// Process does something with the Event: filter, redaction,
	// marshalling, persisting.
	Process(ctx context.Context, e *Event) (*Event, error)
	// Reopen is used to re-read any config stored externally
	// and to close and reopen files, e.g. for log rotation.
	Reopen() error
	// Name returns the node's name.  Nothing enforces uniqueness,
	// but it's usually a good idea.
	Name() string
	// Type describes the type of the node.  This is mostly just used to
	// validate that pipelines are sensibly arranged, e.g. ending with a sink.
	Type() NodeType
}

// A LinkableNode is a Node that has downstream children.  Nodes
// that are *not* LinkableNodes are Leafs.
type LinkableNode interface {
	Node
	SetNext([]Node)
	Next() []Node
}

// LinkNodes is a convenience function that connects
// Nodes together into a linked list. All of the nodes except the
// last one must be LinkableNodes
func LinkNodes(nodes []Node) ([]Node, error) {
	num := len(nodes)
	if num < 2 {
		return nodes, nil
	}

	for i := 0; i < num-1; i++ {
		ln, ok := nodes[i].(LinkableNode)
		if !ok {
			return nil, errors.New("Node is not Linkable")
		}
		ln.SetNext([]Node{nodes[i+1]})
	}

	return nodes, nil
}

// LinkNodesAndSinks is a convenience function that connects
// the inner Nodes together into a linked list.  Then it appends the sinks
// to the end as a set of fan-out leaves.
func LinkNodesAndSinks(inner, sinks []Node) ([]Node, error) {
	_, err := LinkNodes(inner)
	if err != nil {
		return nil, err
	}

	ln, ok := inner[len(inner)-1].(LinkableNode)
	if !ok {
		return nil, fmt.Errorf("last inner node not linkable")
	}

	ln.SetNext(sinks)

	return inner, nil
}
