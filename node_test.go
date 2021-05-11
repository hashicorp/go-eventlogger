package eventlogger

import (
	"testing"

	"github.com/go-test/deep"
)

func TestLinkNodes(t *testing.T) {
	n1, n2, n3 := &Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"}
	root, err := linkNodes([]Node{n1, n2, n3}, []NodeID{"1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}

	expected := &linkedNode{
		node:   n1,
		nodeID: "1",
		next: []*linkedNode{{
			node:   n2,
			nodeID: "2",
			next: []*linkedNode{{
				node:   n3,
				nodeID: "3",
			}},
		}},
	}

	if diff := deep.Equal(root, expected); len(diff) > 0 {
		t.Fatal(diff)
	}
}
