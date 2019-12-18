package eventlogger

import (
	"testing"

	"github.com/go-test/deep"
)

func TestFilter(t *testing.T) {

	predicate := func(e *Envelope) (bool, error) {
		return true, nil
	}
	f := &Filter{Predicate: predicate}

	e := &Envelope{}
	_, err := f.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestByteWriter(t *testing.T) {

	marshaller := func(e *Envelope) ([]byte, error) {
		return make([]byte, 0), nil
	}
	w := &ByteWriter{Marshaller: marshaller}

	e := &Envelope{}
	_, err := w.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

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
	for _, tt := range tests {
		if diff := deep.Equal(tt.linked.Next(), tt.next); diff != nil {
			t.Fatal(diff)
		}
	}
}
