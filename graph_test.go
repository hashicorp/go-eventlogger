package eventlogger

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type reopenNode struct {
	nodes    []Node
	reopened int
}

var _ LinkableNode = &reopenNode{}

func (r *reopenNode) Process(ctx context.Context, e *Event) (*Event, error) {
	return e, nil
}

func (r *reopenNode) SetNext(nodes []Node) {
	r.nodes = nodes
}

func (r *reopenNode) Next() []Node {
	return r.nodes
}

func (r *reopenNode) Reopen() error {
	r.reopened++
	return nil
}

func (r *reopenNode) Type() NodeType {
	return 0
}

func (r *reopenNode) Name() string {
	return "reopenNode"
}

func TestReopen(t *testing.T) {
	nodes, err := LinkNodes([]Node{
		&reopenNode{},
		&reopenNode{},
	})
	if err != nil {
		t.Fatal(err)
	}

	g := graph{
		roots: map[PipelineID]Node{
			"id": nodes[0],
		},
	}
	err = g.reopen(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range nodes {
		if node.(*reopenNode).reopened == 0 {
			t.Fatal("expected node to be reopened")
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

			g := graph{
				roots: map[PipelineID]Node{
					"id": nodes[0],
				},
			}
			err = g.validate()
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
	badsink := &FileSink{Path: "/"}

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

			g := graph{
				roots: map[PipelineID]Node{
					"id": nodes[0],
				},
				successThreshold: tc.threshold,
			}

			err = g.validate()
			if err != nil {
				t.Fatal(err)
			}

			e := &Event{
				Formatted: make(map[string][]byte),
			}
			status, err := g.process(context.Background(), e)
			failure := err != nil
			if failure != tc.failure {
				t.Fatalf("got=%v, expected=%v, error=%v", failure, tc.failure, err)
			}

			if len(status.complete) != tc.sent {
				t.Fatalf("got=%d, expected=%d", len(status.complete), tc.sent)
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

func (fsd *fileSinkDelayed) Process(ctx context.Context, e *Event) (*Event, error) {
	time.Sleep(fsd.delay)
	return fsd.FileSink.Process(ctx, e)
}

func TestSendBlocking(t *testing.T) {
	tmp, err := ioutil.TempFile("", "file.sink.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	goodsink := &FileSink{Path: tmp.Name()}
	slowsink := &fileSinkDelayed{goodsink, time.Second}

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
			[]Node{slowsink}, 0, 0, 0, false,
		},
		{
			"one good no threshold",
			[]Node{goodsink}, 0, 0, 1, false,
		},
		{
			"one good one bad no threshold",
			[]Node{goodsink, slowsink}, 0, 0, 1, false,
		},
		{
			"one bad threshold=1",
			[]Node{slowsink}, 1, 0, 0, true,
		},
		{
			"one good threshold=1",
			[]Node{goodsink}, 1, 0, 1, false,
		},
		{
			"one good one bad threshold=1",
			[]Node{goodsink, slowsink}, 1, 0, 1, false,
		},
		{
			"two bad threshold=2",
			[]Node{slowsink, slowsink}, 2, 0, 0, true,
		},
		{
			"two good threshold=2",
			[]Node{goodsink, goodsink}, 2, 0, 2, false,
		},
		{
			"one good one bad threshold=2",
			[]Node{goodsink, slowsink}, 2, 0, 1, true,
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

			g := graph{
				roots: map[PipelineID]Node{
					"id": nodes[0],
				},
				successThreshold: tc.threshold,
			}

			err = g.validate()
			if err != nil {
				t.Fatal(err)
			}

			e := &Event{
				Formatted: make(map[string][]byte),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			status, err := g.process(ctx, e)
			cancel()
			failure := err != nil
			if failure != tc.failure {
				t.Fatalf("got=%v, expected=%v, error=%v", failure, tc.failure, err)
			}

			if len(status.complete) != tc.sent {
				t.Fatalf("got=%d, expected=%d", len(status.complete), tc.sent)
			}

			if len(status.Warnings) != tc.warnings {
				t.Fatalf("got=%d, expected=%d", len(status.Warnings), tc.warnings)
			}
		})
	}

	// Sleep long enough that the 1s sleep in fileSinkDelayed completes, to
	// satisfy the go leak detector in TestMain.
	time.Sleep(700 * time.Millisecond)
}
