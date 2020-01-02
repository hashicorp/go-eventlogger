package eventlogger

import (
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

	// Construct a Graph
	nodes, err := LinkNodes([]Node{
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
	graph := &Graph{
		Root: nodes[0],
	}
	et := EventType("Foo")
	now := time.Now()
	broker := &Broker{
		Graphs: map[EventType]*Graph{
			et: graph,
		},
		clock: &clock{now},
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
		_, err = broker.Send(nil, et, p)
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
