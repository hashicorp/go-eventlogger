package eventlogger

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-test/deep"
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
			tmpDir, err := ioutil.TempDir("", tt.name)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Send to FileSink
			sink := &FileSink{Path: tmpDir, FileName: "file.log"}
			tt.nodes = append(tt.nodes, sink)

			// Register the graph with the broker
			et := EventType("Foo")
			nodeIDs := nodesToNodeIDs(t, broker, tt.nodes...)
			err = broker.RegisterPipeline(Pipeline{
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
			files, err := ioutil.ReadDir(tmpDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(files) > 1 {
				t.Errorf("Expected 1 log file, got %d", len(files))
			}

			dat, err := ioutil.ReadFile(filepath.Join(tmpDir, files[0].Name()))
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

type testSink struct {
	count int
}

var _ Node = &testSink{}

func (ts *testSink) Type() NodeType {
	return NodeTypeSink
}

func (ts *testSink) Process(ctx context.Context, e *Event) (*Event, error) {
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
		t.Fatalf("expected successThreshold %d, not %d", g.successThreshold, 2)
	}

	err = b.SetSuccessThreshold("t", -1)
	if err == nil || err.Error() != "successThreshold must be 0 or greater" {
		t.Fatalf("expected successThreshold error")
	}
}
