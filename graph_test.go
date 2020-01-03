package eventlogger

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
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

func TestSendResult(t *testing.T) {
	tmp, err := ioutil.TempFile("", "file.sink.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	goodsink := &FileSink{Path: tmp.Name()}
	badsink := &FileSink{Path: "/path/to/file"}

	testcases := []struct {
		name      string
		sinks     []Node
		threshold int
		warnings  int
		sent      int
		failure   bool
	}{
		{
			"one bad no threshold",
			[]Node{badsink}, 0, 1, 0, false,
		},
		{
			"one good no threshold",
			[]Node{goodsink}, 0, 0, 1, false,
		},
		{
			"one good one bad no threshold",
			[]Node{goodsink, badsink}, 0, 1, 1, false,
		},
		{
			"one bad threshold=1",
			[]Node{badsink}, 1, 1, 0, true,
		},
		{
			"one good threshold=1",
			[]Node{goodsink}, 1, 0, 1, false,
		},
		{
			"one good one bad threshold=1",
			[]Node{goodsink, badsink}, 1, 1, 1, false,
		},
		{
			"two bad threshold=2",
			[]Node{badsink, badsink}, 2, 2, 0, true,
		},
		{
			"two good threshold=2",
			[]Node{goodsink, goodsink}, 2, 0, 2, false,
		},
		{
			"one good one bad threshold=2",
			[]Node{goodsink, badsink}, 2, 1, 1, true,
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

			g := Graph{Root: nodes[0], SuccessThreshold: tc.threshold}
			err = g.Validate()
			if err != nil {
				t.Fatal(err)
			}

			e := &Event{
				Formatted: make(map[string][]byte),
			}
			status, err := g.Process(context.Background(), e)
			failure := err != nil
			if failure != tc.failure {
				t.Fatalf("got=%v, expected=%v, error=%v", failure, tc.failure, err)
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

type fileSinkDelayed struct {
	*FileSink
	delay time.Duration
}

var _ Node = &fileSinkDelayed{}

func (fsd *fileSinkDelayed) Process(e *Event) (*Event, error) {
	time.Sleep(fsd.delay)
	return fsd.FileSink.Process(e)
}

func TestSendBlocking(t *testing.T) {
	tmp, err := ioutil.TempFile("", "file.sink.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	goodsink := &FileSink{Path: tmp.Name()}
	badsink := &fileSinkDelayed{goodsink, time.Second}

	// TODO right now we don't flag failure to deliver due to timeout with a
	// warning.  We probably should, in which case maybe we could
	// fold this test into the preceding one.
	testcases := []struct {
		name      string
		sinks     []Node
		threshold int
		warnings  int
		sent      int
		failure   bool
	}{
		{
			"one bad no threshold",
			[]Node{badsink}, 0, 0, 0, false,
		},
		{
			"one good no threshold",
			[]Node{goodsink}, 0, 0, 1, false,
		},
		{
			"one good one bad no threshold",
			[]Node{goodsink, badsink}, 0, 0, 1, false,
		},
		{
			"one bad threshold=1",
			[]Node{badsink}, 1, 0, 0, true,
		},
		{
			"one good threshold=1",
			[]Node{goodsink}, 1, 0, 1, false,
		},
		{
			"one good one bad threshold=1",
			[]Node{goodsink, badsink}, 1, 0, 1, false,
		},
		{
			"two bad threshold=2",
			[]Node{badsink, badsink}, 2, 0, 0, true,
		},
		{
			"two good threshold=2",
			[]Node{goodsink, goodsink}, 2, 0, 2, false,
		},
		{
			"one good one bad threshold=2",
			[]Node{goodsink, badsink}, 2, 0, 1, true,
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

			g := Graph{Root: nodes[0], SuccessThreshold: tc.threshold}
			err = g.Validate()
			if err != nil {
				t.Fatal(err)
			}

			e := &Event{
				Formatted: make(map[string][]byte),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			status, err := g.Process(ctx, e)
			cancel()
			failure := err != nil
			if failure != tc.failure {
				t.Fatalf("got=%v, expected=%v, error=%v", failure, tc.failure, err)
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
