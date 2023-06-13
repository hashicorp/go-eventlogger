// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/hashicorp/go-uuid"
)

func nodesToNodeIDs(t *testing.T, broker *Broker, nodes ...Node) []NodeID {
	t.Helper()
	nodeIDs := make([]NodeID, len(nodes))
	for i, node := range nodes {
		id := NodeID(fmt.Sprintf("node-%d", i))
		err := broker.RegisterNode(id, node)
		if err != nil {
			t.Fatal(err)
		}
		nodeIDs[i] = id
	}
	return nodeIDs
}

func TestBroker(t *testing.T) {
	// Filter out the purple nodes
	filter := &Filter{
		Predicate: func(e *Event) (bool, error) {
			color, ok := e.Payload.(map[string]interface{})["color"]
			return !ok || color != "purple", nil
		},
	}
	// Marshal to JSON
	formatter := &JSONFormatter{}

	formatterFilter := &JSONFormatterFilter{
		Predicate: func(e interface{}) (bool, error) {
			color, ok := e.(*Event).Payload.(map[string]interface{})["color"]
			return !ok || color != "purple", nil
		},
	}

	// Create a broker
	broker := NewBroker()
	now := time.Now()
	broker.clock = &clock{now}

	tests := []struct {
		name  string
		nodes []Node
	}{
		{
			name:  "with-formatter",
			nodes: []Node{filter, formatter},
		},
		{
			name:  "with-formatter-filter",
			nodes: []Node{formatterFilter},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Send to FileSink
			sink := &FileSink{Path: tmpDir, FileName: "file.log"}
			tt.nodes = append(tt.nodes, sink)

			// Register the graph with the broker
			et := EventType("Foo")
			nodeIDs := nodesToNodeIDs(t, broker, tt.nodes...)
			err := broker.RegisterPipeline(Pipeline{
				EventType:  et,
				PipelineID: "id",
				NodeIDs:    nodeIDs,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Set success threshold to 1
			err = broker.SetSuccessThreshold(et, 1)
			if err != nil {
				t.Fatal(err)
			}

			// Process some Events
			payloads := []interface{}{
				map[string]interface{}{
					"color": "red",
					"width": 1,
				},
				map[string]interface{}{
					"color": "green",
					"width": 2,
				},
				map[string]interface{}{
					"color": "purple",
					"width": 3,
				},
				map[string]interface{}{
					"color": "blue",
					"width": 4,
				},
			}
			for _, p := range payloads {
				_, err = broker.Send(context.Background(), et, p)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Check the contents of the log
			dirEntry, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(dirEntry) > 1 {
				t.Errorf("Expected 1 log file, got %d", len(dirEntry))
			}

			dat, err := os.ReadFile(filepath.Join(tmpDir, dirEntry[0].Name()))
			if err != nil {
				t.Fatal(err)
			}

			prefix := fmt.Sprintf(`{"created_at":"%s","event_type":"Foo","payload":`, now.Format(time.RFC3339Nano))
			suffix := "}\n"
			var expect string
			for _, s := range []string{`{"color":"red","width":1}`, `{"color":"green","width":2}`, `{"color":"blue","width":4}`} {
				expect += fmt.Sprintf("%s%s%s", prefix, s, suffix)
			}
			if diff := deep.Equal(string(dat), expect); diff != nil {
				t.Fatal(diff)
			}
		})
	}
}

func TestPipeline(t *testing.T) {
	broker := NewBroker()

	// invalid pipeline
	nodeIDs := nodesToNodeIDs(t, broker, &Filter{Predicate: nil})
	err := broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "id",
		NodeIDs:    nodeIDs,
	})
	if err == nil {
		t.Fatal(err)
	}
	if diff := deep.Equal("non-sink node has no children", err.Error()); diff != nil {
		t.Fatal(diff)
	}

	// Construct a graph
	f1 := &JSONFormatter{}
	s1 := &testSink{}
	p1 := nodesToNodeIDs(t, broker, f1, s1)
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "s1",
		NodeIDs:    p1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// register again
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "s1",
		NodeIDs:    p1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// send a payload
	payload := map[string]interface{}{
		"color": "red",
		"width": 1,
	}
	_, err = broker.Send(context.Background(), "t", payload)
	if err != nil {
		t.Fatal(err)
	}
	if s1.count != 1 {
		t.Fatalf("expected count %d, not %d", s1.count, 1)
	}

	// Construct another graph
	s2 := &testSink{}
	p2 := nodesToNodeIDs(t, broker, f1, s2)
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "s2",
		NodeIDs:    p2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// send a payload
	_, err = broker.Send(context.Background(), "t", payload)
	if err != nil {
		t.Fatal(err)
	}
	if s1.count != 2 {
		t.Fatalf("expected count %d, not %d", s1.count, 2)
	}
	if s2.count != 1 {
		t.Fatalf("expected count %d, not %d", s2.count, 1)
	}

	// remove second graph
	err = broker.RemovePipeline("t", "s2")
	if err != nil {
		t.Fatal(err)
	}

	// send a payload
	_, err = broker.Send(context.Background(), "t", payload)
	if err != nil {
		t.Fatal(err)
	}
	if s1.count != 3 {
		t.Fatalf("expected count %d, not %d", s1.count, 3)
	}
	if s2.count != 1 {
		t.Fatalf("expected count %d, not %d", s2.count, 1)
	}

	// remove
	err = broker.RemovePipeline("t", "s1")
	if err != nil {
		t.Fatal(err)
	}

	// remove again
	err = broker.RemovePipeline("t", "s1")
	if err != nil {
		t.Fatal(err)
	}
}

// TestPipelineRaceCondition can't fail, but it can check if there is a race condition in iterating through, adding, or removing pipelines.
func TestPipelineRaceCondition(t *testing.T) {
	broker := NewBroker()

	eventType := EventType("t")
	var pipelines []PipelineID
	var sinks []*testSink

	wg := sync.WaitGroup{}
	wg.Add(2)

	// register a bunch of pipelines
	go func(t *testing.T) {
		for i := 0; i < 10; i++ {
			// Construct a graph
			f1 := &JSONFormatter{}
			s1 := &testSink{}
			p1 := nodesToNodeIDs(t, broker, f1, s1)

			id, err := uuid.GenerateUUID()
			if err != nil {
				panic(err)
			}

			err = broker.RegisterPipeline(Pipeline{
				EventType:  eventType,
				PipelineID: PipelineID(id),
				NodeIDs:    p1,
			})
			if err != nil {
				panic(err)
			}

			pipelines = append(pipelines, PipelineID(id))
			sinks = append(sinks, s1)
		}

		for _, id := range pipelines {
			err := broker.RemovePipeline(eventType, id)
			if err != nil {
				panic(err)
			}
		}
		wg.Done()
	}(t)

	go func() {
		for i := 0; i < 100; i++ {
			// send payloads
			payload := map[string]interface{}{
				"color": "red",
				"width": 1,
			}
			_, _ = broker.Send(context.Background(), eventType, payload)
		}
		wg.Done()
	}()

	wg.Wait()
}

type testSink struct {
	count int
}

var _ Node = &testSink{}

func (ts *testSink) Type() NodeType {
	return NodeTypeSink
}

func (ts *testSink) Process(_ context.Context, _ *Event) (*Event, error) {
	ts.count++
	return nil, nil
}

func (ts *testSink) Reopen() error {
	return nil
}

func (ts *testSink) Name() string {
	return "testSink"
}

func TestSuccessThreshold(t *testing.T) {
	b := NewBroker()

	err := b.SetSuccessThreshold("t", 2)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := b.graphs["t"]
	if !ok {
		t.Fatalf("expected graph for eventType")
	}
	if g.successThreshold != 2 {
		t.Fatalf("expected successThreshold %d, got %d", 2, g.successThreshold)
	}

	err = b.SetSuccessThreshold("t", -1)
	if err == nil || err.Error() != "successThreshold must be 0 or greater" {
		t.Fatalf("expected successThreshold error")
	}
}

func TestSuccessThresholdSinks(t *testing.T) {
	b := NewBroker()

	err := b.SetSuccessThresholdSinks("t", 2)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := b.graphs["t"]
	if !ok {
		t.Fatalf("expected graph for eventType")
	}
	if g.successThresholdSinks != 2 {
		t.Fatalf("expected successThresholdSinks %d, got %d", 2, g.successThresholdSinks)
	}

	err = b.SetSuccessThresholdSinks("t", -1)
	if err == nil || err.Error() != "successThresholdSinks must be 0 or greater" {
		t.Fatalf("expected successThresholdSinks error")
	}
}

func TestRemovePipelineAndNodes(t *testing.T) {
	broker := NewBroker()

	// Construct a graph
	f1 := &JSONFormatter{}
	s1 := &testSink{}
	nodeIDs := nodesToNodeIDs(t, broker, f1, s1)

	// Register single pipeline
	err := broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p1",
		NodeIDs:    nodeIDs,
	})
	require.NoError(t, err)

	// Deregister the only pipeline we have
	err = broker.RemovePipelineAndNodes(EventType("t"), PipelineID("p1"))
	require.NoError(t, err)
	require.Empty(t, broker.nodes)

	// Attempt to register 2nd pipeline which references now deleted nodes
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p2",
		NodeIDs:    nodeIDs,
	})
	require.Error(t, err)
	require.EqualError(t, err, "nodeID \"node-0\" not registered")

	// Re-register nodes and 2 pipelines
	nodeIDs = nodesToNodeIDs(t, broker, f1, s1)
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p1",
		NodeIDs:    nodeIDs,
	})
	require.NoError(t, err)
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p2",
		NodeIDs:    nodeIDs,
	})
	require.NoError(t, err)

	// Deregister pipeline p1, leave pipeline p2 alone
	err = broker.RemovePipelineAndNodes(EventType("t"), PipelineID("p1"))
	require.NoError(t, err)
	require.NotEmpty(t, broker.nodes)
	require.Equal(t, 2, len(broker.nodes))
}
