package eventlogger

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestBroker(t *testing.T) {

	tmp, err := ioutil.TempFile("", "file.sink.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	path := tmp.Name()

	// Construct a graph
	root, err := linkNodes([]Node{
		// Filter out the purple nodes
		&Filter{
			Predicate: func(e *Event) (bool, error) {
				color, ok := e.Payload.(map[string]interface{})["color"]
				return !ok || color != "purple", nil
			},
		},
		// Marshal to JSON
		&JSONFormatter{},
		// Send to FileSink
		&FileSink{
			Path: path,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a broker
	broker := NewBroker()
	now := time.Now()
	broker.clock = &clock{now}

	// Register the graph with the broker
	et := EventType("Foo")
	err = broker.RegisterPipeline(et, "id", root)
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
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	prefix := fmt.Sprintf(`{"CreatedAt":"%s","EventType":"Foo","Payload":`, now.Format(time.RFC3339Nano))
	suffix := "}\n"
	var expect string
	for _, s := range []string{`{"color":"red","width":1}`, `{"color":"green","width":2}`, `{"color":"blue","width":4}`} {
		expect += fmt.Sprintf("%s%s%s", prefix, s, suffix)
	}
	if diff := deep.Equal(string(dat), expect); diff != nil {
		t.Fatal(diff)
	}
}

func TestPipeline(t *testing.T) {
	broker := NewBroker()

	// invalid pipeline
	root := &linkedNode{node: &Filter{Predicate: nil}}
	err := broker.RegisterPipeline("t", "id", root)
	if err == nil {
		t.Fatal(err)
	}
	if diff := deep.Equal("non-sink node has no children", err.Error()); diff != nil {
		t.Fatal(diff)
	}

	// Construct a graph
	s1 := &testSink{}
	p1, err := linkNodes([]Node{
		&JSONFormatter{},
		s1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// register
	err = broker.RegisterPipeline("t", "s1", p1)
	if err != nil {
		t.Fatal(err)
	}

	// register again
	err = broker.RegisterPipeline("t", "s1", p1)
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
	p2, err := linkNodes([]Node{
		&JSONFormatter{},
		s2,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = broker.RegisterPipeline("t", "s2", p2)
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
