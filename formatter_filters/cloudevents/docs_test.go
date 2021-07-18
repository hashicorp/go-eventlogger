package cloudevents_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/eventlogger/formatter_filters/cloudevents"
	"github.com/hashicorp/eventlogger/sinks/writer"
)

func ExampleFormatter() {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	// Create a broker
	b := eventlogger.NewBroker()

	b.StopTimeAt(then) // setting this so the output timestamps are predictable for testing.

	eventSource, err := url.Parse("https://github.com/hashicorp/go-eventlogger/formatters/cloudevents")
	if err != nil {
		// handle err
	}
	// format as cloudevents and filter "admin" roles
	cloudEventsFmt := &cloudevents.FormatterFilter{
		Format: cloudevents.FormatJSON,
		Source: eventSource,
		Predicate: func(ctx context.Context, ce interface{}) (bool, error) {
			role, ok := ce.(cloudevents.Event).Data.(map[string]interface{})["role"]
			return !ok || role != "admin", nil
		},
	}

	// Send the output to stdout
	stdoutSink := &writer.Sink{
		Writer: os.Stdout,
		Format: string(cloudevents.FormatJSON),
	}

	// Register the nodes with the broker
	nodes := []eventlogger.Node{cloudEventsFmt, stdoutSink}
	nodeIDs := make([]eventlogger.NodeID, len(nodes))
	for i, node := range nodes {
		id := eventlogger.NodeID(fmt.Sprintf("node-%d", i))
		err := b.RegisterNode(id, node)
		if err != nil {
			// handle error
		}
		nodeIDs[i] = id
	}

	et := eventlogger.EventType("test-event")
	// Register a pipeline for our event type
	err = b.RegisterPipeline(eventlogger.Pipeline{
		EventType:  et,
		PipelineID: "cloudevents-pipeline",
		NodeIDs:    nodeIDs,
	})
	if err != nil {
		// handle error
	}

	payloads := []*testPayload{
		{
			payload: map[string]interface{}{
				"id": "test-id",
				"data": map[string]interface{}{
					"name":      "bob",
					"role":      "user",
					"pronouns":  []string{"they", "them"},
					"coworkers": []string{"alice", "eve"},
				},
			},
		},
		{
			payload: map[string]interface{}{
				"id": "test-id",
				"data": map[string]interface{}{
					"name": "alice",
					"role": "admin",
				},
			},
		},
	}
	for _, p := range payloads {
		if status, err := b.Send(context.Background(), et, p); err != nil {
			// handle err and status.Warnings
			fmt.Println("err: ", err)
			fmt.Println("warnings: ", status.Warnings)
		}
	}

	// Output:
	// {"id":"test-id","source":"https://github.com/hashicorp/go-eventlogger/formatters/cloudevents","specversion":"1.0","type":"test-event","data":{"coworkers":["alice","eve"],"name":"bob","pronouns":["they","them"],"role":"user"},"datacontentype":"application/cloudevents","time":"2009-11-17T20:34:58.651387237Z"}
}

type testPayload struct {
	payload map[string]interface{}
}

func (t *testPayload) ID() string {
	return t.payload["id"].(string)
}

func (t *testPayload) Data() interface{} {
	return t.payload["data"].(interface{})
}
