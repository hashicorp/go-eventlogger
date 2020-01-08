package eventlogger

import (
	"testing"

	"github.com/go-test/deep"
)

func TestLinkNodes(t *testing.T) {
	nodes, err := LinkNodes([]Node{
		&Filter{Predicate: nil},
		&JSONFormatter{},
		&FileSink{Path: "test.log"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		linked LinkableNode
		next   []Node
	}{
		{nodes[0].(LinkableNode), []Node{nodes[1]}},
		{nodes[1].(LinkableNode), []Node{nodes[2]}},
	}
	for _, tt := range tests {
		if diff := deep.Equal(tt.linked.Next(), tt.next); diff != nil {
			t.Fatal(diff)
		}
	}
}
