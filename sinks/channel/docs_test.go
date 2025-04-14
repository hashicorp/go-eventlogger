// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel_test

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/eventlogger/sinks/channel"
)

func ExampleChannelSink() {
	then := time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	// Create a broker
	b, _ := eventlogger.NewBroker()

	b.StopTimeAt(then) // setting this so the output timestamps are predictable for testing.

	// Marshal to JSON
	jsonFmt := &jsonfmt.JSONFormatter{}

	// Send the output to a channel
	testChan := make(chan *eventlogger.Event, 1)
	chanSink, err := channel.NewChannelSink(testChan, time.Second)
	if err != nil {
		// handle error
	}

	// Register the nodes with the broker
	nodes := []eventlogger.Node{jsonFmt, chanSink}
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
		PipelineID: "writer-sink-pipeline",
		NodeIDs:    nodeIDs,
	})
	if err != nil {
		// handle error
	}

	p := map[string]interface{}{
		"name":      "bob",
		"role":      "user",
		"pronouns":  []string{"they", "them"},
		"coworkers": []string{"alice", "eve"},
	}
	// Send an event
	if status, err := b.Send(context.Background(), et, p); err != nil {
		// handle err and status.Warnings
		fmt.Println("err: ", err)
		fmt.Println("warnings: ", status.Warnings)
	}

	output := <-testChan
	fmt.Println(string(output.Formatted["json"]))

	// Output:
	// {"created_at":"2009-11-17T20:34:58.651387237Z","event_type":"test-event","payload":{"coworkers":["alice","eve"],"name":"bob","pronouns":["they","them"],"role":"user"}}
}
