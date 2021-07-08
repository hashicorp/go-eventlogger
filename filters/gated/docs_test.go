package gated_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/eventlogger/filters/gated"
	"github.com/hashicorp/eventlogger/sinks/writer"
)

func ExampleFilter() {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	// Create a broker
	b := eventlogger.NewBroker()

	b.StopTimeAt(then) // setting this so the output timestamps are predictable for testing.

	// A gated.Filter for events
	gf := &gated.Filter{
		Broker:  b,
		NowFunc: func() time.Time { return then }, // setting this so the output timestamps are predictable for testing.
	}
	// Marshal to JSON
	jsonFmt := &eventlogger.JSONFormatter{}

	// Send the output to stdout
	stdoutSink := &writer.Sink{
		Writer: os.Stdout,
	}

	// Register the nodes with the broker
	nodes := []eventlogger.Node{gf, jsonFmt, stdoutSink}
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
	err := b.RegisterPipeline(eventlogger.Pipeline{
		EventType:  et,
		PipelineID: "gated-filter-pipeline",
		NodeIDs:    nodeIDs,
	})
	if err != nil {
		// handle error
	}

	// define a common event ID for a set of events we want gated together.
	eventID := "event-1"
	payloads := []*gated.Payload{
		{
			// our first event
			ID: eventID,
			Header: map[string]interface{}{
				"tmz":  "EST",
				"user": "alice",
			},
			Detail: map[string]interface{}{
				"file_name":   "file1.txt",
				"total_bytes": 1024,
			},
		},
		{
			// our 2nd event
			ID: eventID,
			Header: map[string]interface{}{
				"roles": []string{"admin", "individual-contributor"},
			},
		},
		// the last event
		{
			ID:    eventID,
			Flush: true,
			Detail: map[string]interface{}{
				"file_name":   "file2.txt",
				"total_bytes": 512,
			},
		},
	}

	ctx := context.Background()
	for _, p := range payloads {
		// Send our gated event payloads
		if status, err := b.Send(ctx, et, p); err != nil {
			// handle err and status.Warnings
			fmt.Println("err: ", err)
			fmt.Println("warnings: ", status.Warnings)
		}
	}

	// Output:
	// {"created_at":"2009-11-17T20:34:58.651387237Z","event_type":"test-event","payload":{"id":"event-1","header":{"roles":["admin","individual-contributor"],"tmz":"EST","user":"alice"},"details":[{"type":"test-event","created_at":"2009-11-17 20:34:58.651387237 +0000 UTC","payload":{"file_name":"file1.txt","total_bytes":1024}},{"type":"test-event","created_at":"2009-11-17 20:34:58.651387237 +0000 UTC","payload":{"file_name":"file2.txt","total_bytes":512}}]}}
}
