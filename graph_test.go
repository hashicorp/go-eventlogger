package eventlogger

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestGraph(t *testing.T) {

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
				color, ok := e.Payload["color"]
				return !ok || color != "purple", nil
			},
		},
		// Marshal to JSON
		&ByteWriter{
			Marshaller: JSONMarshaller,
		},
		// Send to FileSink
		&FileSink{
			FilePath: path,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	graph := &Graph{
		Root: nodes[0],
	}

	// Process some Events
	envelopes := []*Event{
		&Event{
			Payload: map[string]interface{}{
				"color": "red",
				"width": 1,
			},
		},
		&Event{
			Payload: map[string]interface{}{
				"color": "green",
				"width": 2,
			},
		},
		&Event{
			Payload: map[string]interface{}{
				"color": "purple",
				"width": 3,
			},
		},
		&Event{
			Payload: map[string]interface{}{
				"color": "blue",
				"width": 4,
			},
		},
	}
	for _, env := range envelopes {
		err = graph.Process(env)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Check the contents of the log
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expect := `{"color":"red","width":1}
{"color":"green","width":2}
{"color":"blue","width":4}
`
	if diff := deep.Equal(string(dat), expect); diff != nil {
		t.Fatal(diff)
	}
}
