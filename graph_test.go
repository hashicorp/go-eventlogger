package eventlogger

import (
	"context"
	"testing"
)

type reloadNode struct {
	nodes    []Node
	reloaded int
}

var _ LinkableNode = &reloadNode{}

func (r *reloadNode) Process(e *Event) (*Event, error) {
	return e, nil
}

func (r *reloadNode) SetNext(nodes []Node) {
	r.nodes = nodes
}

func (r *reloadNode) Next() []Node {
	return r.nodes
}

func (r *reloadNode) Reload() error {
	r.reloaded++
	return nil
}

func (r *reloadNode) Type() NodeType {
	return 0
}

func TestReload(t *testing.T) {
	nodes, err := LinkNodes([]Node{
		&reloadNode{},
		&reloadNode{},
	})
	if err != nil {
		t.Fatal(err)
	}

	g := Graph{Root: nodes[0]}
	err = g.Reload(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range nodes {
		if node.(*reloadNode).reloaded == 0 {
			t.Fatal("expected node to be reloaded")
		}
	}
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		name  string
		nodes []Node
		valid bool
	}{
		{
			"childless inner node",
			[]Node{
				&ByteWriter{
					Marshaller: JSONMarshaller,
				},
			},
			false,
		},
		{
			"root sink",
			[]Node{
				&FileSink{
					Path: "/path/to/file",
				},
			},
			false,
		},
		{
			"sink without formatter",
			[]Node{
				&Filter{
					Predicate: func(e *Event) (bool, error) { return true, nil },
				},
				&FileSink{
					Path: "/path/to/file",
				},
			},
			false,
		},
		{
			"good graph",
			[]Node{
				&ByteWriter{
					Marshaller: JSONMarshaller,
				},
				&FileSink{
					Path: "/path/to/file",
				},
			},
			true,
		},
	}

	for i := range testcases {
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			nodes, err := LinkNodes(tc.nodes)
			if err != nil {
				t.Fatal(err)
			}

			g := Graph{Root: nodes[0]}
			err = g.Validate()
			valid := err == nil
			if valid != tc.valid {
				t.Fatalf("valid=%v, expected=%v, err=%v", valid, tc.valid, err)
			}

		})
	}

}
