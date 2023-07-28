// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type reopenNode struct {
	reopened int
}

var _ Node = &reopenNode{}

func (r *reopenNode) Process(ctx context.Context, e *Event) (*Event, error) {
	return e, nil
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
	nodes := []Node{&reopenNode{}, &reopenNode{}}
	root, err := linkNodes(nodes, []NodeID{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}

	g := graph{}
	g.roots.Store("id", root)
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
			"good graph using a formatter",
			[]Node{
				&JSONFormatter{},
				&FileSink{
					Path: "/path/to/file",
				},
			},
			true,
		},
		{
			"good graph using a formatter filter",
			[]Node{
				&JSONFormatterFilter{},
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
			ids := make([]NodeID, len(tc.nodes))
			root, err := linkNodes(tc.nodes, ids)
			if err != nil {
				t.Fatal(err)
			}

			g := graph{}
			g.roots.Store("id", root)
			err = g.validate()
			valid := err == nil
			if valid != tc.valid {
				t.Fatalf("valid=%v, expected=%v, err=%v", valid, tc.valid, err)
			}
		})
	}
}

func TestSendResult(t *testing.T) {
	tmpDir := t.TempDir()
	goodSink := NodeID("good")
	badSink := NodeID("bad")
	sinksByID := map[NodeID]Node{
		goodSink: &FileSink{Path: tmpDir, FileName: "sink"},
		badSink:  &FileSink{Path: "/"},
	}

	testCases := []struct {
		name           string
		sinkIDs        []NodeID
		threshold      int
		thresholdSinks int
		warnings       int
		sent           int
		completeSinks  int
		failure        bool
		filter         bool
	}{
		{
			name:      "one bad no threshold/thresholdSinks",
			sinkIDs:   []NodeID{badSink},
			threshold: 0, thresholdSinks: 0, warnings: 1, sent: 0, completeSinks: 0, failure: false},
		{
			name:      "one good no threshold/thresholdSinks",
			sinkIDs:   []NodeID{goodSink},
			threshold: 0, thresholdSinks: 0, warnings: 0, sent: 1, completeSinks: 1, failure: false,
		},
		{
			name:      "one good one bad no threshold/thresholdSinks",
			sinkIDs:   []NodeID{goodSink, badSink},
			threshold: 0, thresholdSinks: 0, warnings: 1, sent: 1, completeSinks: 1, failure: false,
		},
		{
			name:      "one bad threshold=1 thresholdSinks=1",
			sinkIDs:   []NodeID{badSink},
			threshold: 1, thresholdSinks: 1, warnings: 1, sent: 0, completeSinks: 0, failure: true,
		},
		{
			name:      "one good threshold=1 thresholdSinks=0",
			sinkIDs:   []NodeID{goodSink},
			threshold: 1, thresholdSinks: 0, warnings: 0, sent: 1, completeSinks: 1, failure: false,
		},
		{
			name:      "one good one bad threshold=1 thresholdSinks=1",
			sinkIDs:   []NodeID{goodSink, badSink},
			threshold: 1, thresholdSinks: 1, warnings: 1, sent: 1, completeSinks: 1, failure: false,
		},
		{
			name:      "two bad threshold=2 thresholdSinks=2",
			sinkIDs:   []NodeID{badSink, badSink},
			threshold: 2, thresholdSinks: 2, warnings: 2, sent: 0, completeSinks: 0, failure: true,
		},
		{
			name:      "two good threshold=2 thresholdSinks=2",
			sinkIDs:   []NodeID{goodSink, goodSink},
			threshold: 2, thresholdSinks: 2, warnings: 0, sent: 2, completeSinks: 2, failure: false,
		},
		{
			name:      "one good one bad threshold=2 thresholdSinks=2",
			sinkIDs:   []NodeID{goodSink, badSink},
			threshold: 2, thresholdSinks: 2, warnings: 1, sent: 1, completeSinks: 1, failure: true,
		},
		{
			name:      "one bad sink with filter threshold=1 thresholdSink=0",
			sinkIDs:   []NodeID{badSink},
			threshold: 1, thresholdSinks: 0, warnings: 0, sent: 1, completeSinks: 0, failure: false, filter: true,
		},
		{
			name:      "one bad sink with filter threshold=1 thresholdSinks=1",
			sinkIDs:   []NodeID{badSink},
			threshold: 1, thresholdSinks: 1, warnings: 0, sent: 1, completeSinks: 0, failure: true, filter: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			predicate := func(e *Event) (bool, error) {
				return !tc.filter, nil
			}
			nodes := []Node{&Filter{Predicate: predicate}, &JSONFormatter{}}
			sinks := make([]Node, len(tc.sinkIDs))
			for i, id := range tc.sinkIDs {
				sinks[i] = sinksByID[id]
			}
			root, err := linkNodesAndSinks(nodes, sinks, []NodeID{"filter1", "formatter1"}, tc.sinkIDs)
			if err != nil {
				t.Fatal(err)
			}

			g := graph{successThreshold: tc.threshold, successThresholdSinks: tc.thresholdSinks}
			g.roots.Store("id", root)

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

			if len(status.completeSinks) != tc.completeSinks {
				t.Fatalf("completeSinks: got=%d, expected=%d", len(status.completeSinks), tc.completeSinks)
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
	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	fs := &FileSink{Path: tmpDir, FileName: "sink"}
	goodsink := NodeID("good")
	slowsink := NodeID("bad")
	sinksByID := map[NodeID]Node{
		goodsink: fs,
		slowsink: &fileSinkDelayed{fs, time.Second},
	}

	// TODO right now we don't flag failure to deliver due to timeout with a
	// warning.  We probably should, in which case maybe we could
	// fold this test into the preceding one.
	testcases := []struct {
		name      string
		sinkIDs   []NodeID
		threshold int
		warnings  int
		sent      int
		failure   bool
	}{
		{
			"one bad no threshold",
			[]NodeID{slowsink},
			0, 0, 0, false,
		},
		{
			"one good no threshold",
			[]NodeID{goodsink},
			0, 0, 1, false,
		},
		{
			"one good one bad no threshold",
			[]NodeID{goodsink, slowsink},
			0, 0, 1, false,
		},
		{
			"one bad threshold=1",
			[]NodeID{slowsink},
			1, 0, 0, true,
		},
		{
			"one good threshold=1",
			[]NodeID{goodsink},
			1, 0, 1, false,
		},
		{
			"one good one bad threshold=1",
			[]NodeID{goodsink, slowsink},
			1, 0, 1, false,
		},
		{
			"two bad threshold=2",
			[]NodeID{slowsink, slowsink},
			2, 0, 0, true,
		},
		{
			"two good threshold=2",
			[]NodeID{goodsink, goodsink},
			2, 0, 2, false,
		},
		{
			"one good one bad threshold=2",
			[]NodeID{goodsink, slowsink},
			2, 0, 1, true,
		},
	}

	for i := range testcases {
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			nodes := []Node{&JSONFormatter{}}
			sinks := make([]Node, len(tc.sinkIDs))
			for i, id := range tc.sinkIDs {
				sinks[i] = sinksByID[id]
			}
			root, err := linkNodesAndSinks(nodes, sinks, []NodeID{"f1"}, tc.sinkIDs)
			if err != nil {
				t.Fatal(err)
			}

			g := graph{successThreshold: tc.threshold}
			g.roots.Store("id", root)

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

// TestGraph_HasRegistrations ensures that hasRegistrations performs as expected.
// If anything has been stored against the roots then we should expect true,
// otherwise false.
func TestGraph_HasRegistrations(t *testing.T) {
	g := graph{}

	// No registrations to begin with.
	require.False(t, g.hasRegistrations())

	// Store something and ensure we have registrations.
	g.roots.Store("abc", nil)
	require.True(t, g.hasRegistrations())

	// Delete the stored entry and ensure we have no registrations.
	g.roots.Delete("abc")
	require.False(t, g.hasRegistrations())
}
