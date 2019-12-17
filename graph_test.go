package eventlogger

import (
	"testing"

	"github.com/go-test/deep"
)

func TestLinkNodes(t *testing.T) {

	nodes, err := LinkNodes([]Node{
		&Filter{Predicate: nil},
		&ByteWriter{Marshaller: nil},
		&FileSink{FilePath: "test.log"},
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
	for i := 0; i < len(tests); i++ {
		if diff := deep.Equal(tests[i].linked.Next(), tests[i].next); diff != nil {
			t.Fatal(diff)
		}
	}
}
