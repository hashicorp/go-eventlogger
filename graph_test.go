package eventlogger

import (
	"context"
	"io/ioutil"
	"os"
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

func (r *reloadNode) Name() string {
	return "reloadNode"
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
				&JSONFormatter{},
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
				&JSONFormatter{},
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

func TestStatus(t *testing.T) {
	tmp, err := ioutil.TempFile("", "file.sink.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	goodpath := tmp.Name()
	badpath := "/"

	testcases := []struct {
		name     string
		sinks    []Node
		warnings int
		sent     int
	}{
		{
			"one bad",
			[]Node{
				&FileSink{
					Path: badpath,
				},
			},
			1,
			0,
		},
		{
			"one good",
			[]Node{
				&FileSink{
					Path: goodpath,
				},
			},
			0,
			1,
		},
		{
			"one good one bad",
			[]Node{
				&FileSink{
					Path: badpath,
				},
				&FileSink{
					Path: goodpath,
				},
			},
			1,
			1,
		},
	}

	for i := range testcases {
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			nodes := []Node{&JSONFormatter{}}
			_, err := LinkNodesAndSinks(nodes, tc.sinks)
			if err != nil {
				t.Fatal(err)
			}

			g := Graph{Root: nodes[0]}
			err = g.Validate()
			if err != nil {
				t.Fatal(err)
			}

			e := &Event{
				Formatted: make(map[string][]byte),
			}
			status, err := g.Process(context.Background(), e)
			if err != nil {
				t.Fatal(err)
			}
			if len(status.SentToSinks) != tc.sent {
				t.Fatalf("got=%d, expected=%d", len(status.SentToSinks), tc.sent)
			}

			if len(status.Warnings) != tc.warnings {
				t.Fatalf("got=%d, expected=%d", len(status.Warnings), tc.warnings)
			}
		})
	}

}
