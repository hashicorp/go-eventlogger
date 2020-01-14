package eventlogger

import (
	"context"
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

type linkedNode struct {
	node Node
	next []*linkedNode
}

// linkNodes is a convenience function that connects
// Nodes together into a linked list. All of the nodes except the
// last one must be LinkableNodes
func linkNodes(nodes []Node) (*linkedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes given")
	}

	root := &linkedNode{node: nodes[0]}
	cur := root

	for _, n := range nodes[1:] {
		next := &linkedNode{node: n}
		cur.next = []*linkedNode{next}
		cur = next
	}

	return root, nil
}

// linkNodesAndSinks is a convenience function that connects
// the inner Nodes together into a linked list.  Then it appends the sinks
// to the end as a set of fan-out leaves.
func linkNodesAndSinks(inner, sinks []Node) (*linkedNode, error) {
	root, err := linkNodes(inner)
	if err != nil {
		return nil, err
	}

	// This is inefficient but since it's only used in setup we don't care:
	cur := root
	for cur.next != nil {
		cur = cur.next[0]
	}

	for _, s := range sinks {
		cur.next = append(cur.next, &linkedNode{node: s})
	}

	return root, nil
}
