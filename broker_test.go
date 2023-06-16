// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-test/deep"
	"github.com/hashicorp/go-uuid"
)

// nodesToNodeIDs takes the supplied nodes and registers them with a corresponding,
// generated ID which follows the format 'node-{argument_index}' starting from 0.
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
	broker, err := NewBroker()
	require.NoError(t, err)
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
	broker, err := NewBroker()
	require.NoError(t, err)

	// invalid pipeline
	nodeIDs := nodesToNodeIDs(t, broker, &Filter{Predicate: nil})
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "id",
		NodeIDs:    nodeIDs,
	})
	require.Error(t, err)

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
	require.NoError(t, err)

	// register again
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "s1",
		NodeIDs:    p1,
	})
	require.NoError(t, err)

	// send a payload
	payload := map[string]interface{}{
		"color": "red",
		"width": 1,
	}
	_, err = broker.Send(context.Background(), "t", payload)
	require.NoError(t, err)
	require.Equal(t, 1, s1.count)

	// Construct another graph
	s2 := &testSink{}
	p2 := nodesToNodeIDs(t, broker, f1, s2)
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "s2",
		NodeIDs:    p2,
	})
	require.NoError(t, err)

	// send a payload
	_, err = broker.Send(context.Background(), "t", payload)
	require.NoError(t, err)
	require.Equal(t, 2, s1.count)
	require.Equal(t, 1, s2.count)

	// remove second graph
	err = broker.RemovePipeline("t", "s2")
	require.NoError(t, err)

	// send a payload
	_, err = broker.Send(context.Background(), "t", payload)
	require.NoError(t, err)
	require.Equal(t, 3, s1.count)
	require.Equal(t, 1, s2.count)

	// remove
	err = broker.RemovePipeline("t", "s1")
	require.NoError(t, err)

	// remove again
	err = broker.RemovePipeline("t", "s1")
	require.NoError(t, err)
}

// TestPipelineRaceCondition can't fail, but it can check if there is a race condition in iterating through, adding, or removing pipelines.
func TestPipelineRaceCondition(t *testing.T) {
	broker, err := NewBroker()
	require.NoError(t, err)

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

// TestSuccessThreshold tests that we can set the required success threshold for
// a specific event type in the graph.
func TestSuccessThreshold(t *testing.T) {
	threshold := 2
	b, err := NewBroker()
	require.NoError(t, err)

	err = b.SetSuccessThreshold("", threshold)
	require.Error(t, err)
	require.EqualError(t, err, "event type cannot be empty")

	err = b.SetSuccessThreshold("t", threshold)
	require.NoError(t, err)

	g, ok := b.graphs["t"]
	require.True(t, ok)
	require.Equal(t, threshold, g.successThreshold)

	err = b.SetSuccessThreshold("t", -1)
	require.Error(t, err)
	require.EqualError(t, err, "successThreshold must be 0 or greater")
}

// TestSuccessThresholdSinks tests that we can set the required sink success
// threshold for a specific event type in the graph.
func TestSuccessThresholdSinks(t *testing.T) {
	threshold := 2
	b, err := NewBroker()
	require.NoError(t, err)

	err = b.SetSuccessThresholdSinks("", threshold)
	require.Error(t, err)
	require.EqualError(t, err, "event type cannot be empty")

	err = b.SetSuccessThresholdSinks("t", threshold)
	require.NoError(t, err)

	g, ok := b.graphs["t"]
	require.True(t, ok)
	require.Equal(t, threshold, g.successThresholdSinks)

	err = b.SetSuccessThresholdSinks("t", -1)
	require.Error(t, err)
	require.EqualError(t, err, "successThresholdSinks must be 0 or greater")
}

// TestRemovePipelineAndNodes exercises the behavior that removes a pipeline and
// any nodes associated with that pipeline, if they are not referenced by other pipelines.
// The test is relatively long as it is focused on the state of the broker across
// multiple operations.
func TestRemovePipelineAndNodes(t *testing.T) {
	broker, err := NewBroker()
	require.NoError(t, err)

	// Construct a graph
	f1 := &JSONFormatter{}
	s1 := &testSink{}
	nodeIDs := nodesToNodeIDs(t, broker, f1, s1)

	// Register single pipeline
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p1",
		NodeIDs:    nodeIDs,
	})
	require.NoError(t, err)

	// Deregister the only pipeline we have
	ok, err := broker.RemovePipelineAndNodes(EventType("t"), PipelineID("p1"))
	require.NoError(t, err)
	require.True(t, ok)
	require.Empty(t, broker.nodes)

	// Attempt to register 2nd pipeline which references now deleted nodes
	err = broker.RegisterPipeline(Pipeline{
		EventType:  "t",
		PipelineID: "p2",
		NodeIDs:    nodeIDs,
	})
	require.Error(t, err)
	require.EqualError(t, err, "node ID \"node-0\" not registered")

	// Re-register nodes and 2 pipelines which use the same nodes
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

	// Deregister pipeline p1, leave pipeline p2 in place
	ok, err = broker.RemovePipelineAndNodes(EventType("t"), PipelineID("p1"))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, broker.nodes)
	require.Equal(t, 2, len(broker.nodes))

	// Whip the nodes out from underneath a pipeline and then try to deregister it
	broker.nodes = nil
	ok, err = broker.RemovePipelineAndNodes(EventType("t"), "p2")
	require.Error(t, err)
	require.True(t, ok)
	me, ok := err.(*multierror.Error)
	require.True(t, ok)
	require.Equal(t, 2, me.Len())
}

// TestRemovePipelineAndNodes_BadEventType tests attempting to remove a pipeline
// with an event type we haven't previously registered.
func TestRemovePipelineAndNodes_BadEventType(t *testing.T) {
	broker, err := NewBroker()
	require.NoError(t, err)
	ok, err := broker.RemovePipelineAndNodes(EventType("foo"), PipelineID("p2"))
	require.Error(t, err)
	require.False(t, ok)
	require.EqualError(t, err, "no graph for EventType foo")
}

// TestRegisterPipeline_BadParameters ensures that we perform sanity checking
// on the parameters passed in when we attempt to register a pipeline.
func TestRegisterPipeline_BadParameters(t *testing.T) {
	tests := map[string]struct {
		pipelineID string
		eventType  string
		nodes      []NodeID
		error      string
	}{
		"no-pipelineID": {
			pipelineID: "",
			eventType:  "t",
			nodes:      []NodeID{"1", "2", "3"},
			error:      "pipeline ID is required",
		},
		"no-eventType": {
			pipelineID: "1",
			eventType:  "",
			nodes:      []NodeID{"1", "2", "3"},
			error:      "event type is required",
		},
		"nil-nodes": {
			pipelineID: "1",
			eventType:  "t",
			nodes:      nil,
			error:      "node IDs are required",
		},
		"empty-nodes": {
			pipelineID: "1",
			eventType:  "t",
			nodes:      []NodeID{},
			error:      "node IDs are required",
		},
		"bad-nodes": {
			pipelineID: "1",
			eventType:  "t",
			nodes:      []NodeID{"", ""},
			error:      "node ID cannot be empty",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			broker, err := NewBroker()
			require.NoError(t, err)

			// Register some nodes so they exist (1, 2, 3 should appear in the test cases)
			node := &JSONFormatter{}
			err = broker.RegisterNode("1", node)
			require.NoError(t, err)
			err = broker.RegisterNode("2", node)
			require.NoError(t, err)
			err = broker.RegisterNode("3", node)
			require.NoError(t, err)

			err = broker.RegisterPipeline(Pipeline{
				PipelineID: PipelineID(tc.pipelineID),
				EventType:  EventType(tc.eventType),
				NodeIDs:    tc.nodes,
			})

			require.Error(t, err)
			me, ok := err.(*multierror.Error)
			require.True(t, ok)
			require.EqualError(t, me.Unwrap(), tc.error)
		})
	}
}

// TestRemovePipelineAndNodes_BadParameters ensures that we perform sanity checking
// on the parameters passed in when we attempt to remove both individual pipelines
// and also pipelines and nodes together.
func TestRemovePipelineAndNodes_BadParameters(t *testing.T) {
	tests := map[string]struct {
		pipelineID string
		eventType  string
		error      string
	}{
		"no-pipelineID": {
			pipelineID: "",
			eventType:  "t",
			error:      "pipeline ID cannot be empty",
		},
		"no-eventType": {
			pipelineID: "1",
			eventType:  "",
			error:      "event type cannot be empty",
		},
		"wrong-eventType": {
			pipelineID: "1",
			eventType:  "foo",
			error:      "no graph for EventType foo",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			broker, err := NewBroker()
			require.NoError(t, err)

			// Test removing the pipeline and nodes
			ok, err := broker.RemovePipelineAndNodes(EventType(tc.eventType), PipelineID(tc.pipelineID))
			require.Error(t, err)
			require.False(t, ok)
			require.EqualError(t, err, tc.error)
			// Test removing just the pipeline
			err = broker.RemovePipeline(EventType(tc.eventType), PipelineID(tc.pipelineID))
			require.Error(t, err)
			require.EqualError(t, err, tc.error)
		})
	}
}

// TestPipelineValidate tests that given a Pipeline in various states we can assert
// that the Pipeline is valid or invalid.
func TestPipelineValidate(t *testing.T) {
	tests := map[string]struct {
		pipelineID       string
		eventType        string
		nodes            []NodeID
		expectValid      bool
		expectErrorCount int
	}{
		"valid": {
			pipelineID:  "1",
			eventType:   "t",
			nodes:       []NodeID{"a", "b"},
			expectValid: true,
		},
		"no-pipelineID": {
			pipelineID:       "",
			eventType:        "t",
			nodes:            []NodeID{"a", "b"},
			expectValid:      false,
			expectErrorCount: 1,
		},
		"no-event-type": {
			pipelineID:       "1",
			eventType:        "",
			nodes:            []NodeID{"a", "b"},
			expectValid:      false,
			expectErrorCount: 1,
		},
		"empty-nodes": {
			pipelineID:       "1",
			eventType:        "t",
			nodes:            []NodeID{},
			expectValid:      false,
			expectErrorCount: 1,
		},
		"fully-invalid": {
			pipelineID:       "",
			eventType:        "",
			nodes:            []NodeID{},
			expectValid:      false,
			expectErrorCount: 3,
		},
		"fully-invalid-no-nodeIDs": {
			pipelineID:       "",
			eventType:        "",
			nodes:            []NodeID{"", ""},
			expectValid:      false,
			expectErrorCount: 3,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := Pipeline{
				PipelineID: PipelineID(tc.pipelineID),
				EventType:  EventType(tc.eventType),
				NodeIDs:    tc.nodes,
			}

			err := p.validate()
			switch tc.expectValid {
			case true:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				me, ok := err.(*multierror.Error)
				require.True(t, ok)
				require.Equal(t, tc.expectErrorCount, me.Len())
			}
		})
	}
}

// TestRegisterNode_NoID ensures we cannot register a Node with an empty ID.
func TestRegisterNode_NoID(t *testing.T) {
	b, err := NewBroker()
	require.NoError(t, err)
	err = b.RegisterNode("", &JSONFormatter{})
	require.Error(t, err)
	require.EqualError(t, err, "unable to register node, node ID cannot be empty")
}

// TestBroker_RegisterNode_AllowOverwrite_Implicit is used to prove that nodes can be
// overwritten when a Broker has been implicitly configured with the AllowOverwrite policy.
// This is the default in order to maintain pre-existing behavior.
func TestBroker_RegisterNode_AllowOverwrite_Implicit(t *testing.T) {
	b, err := NewBroker()
	require.NoError(t, err)
	err = b.RegisterNode("n1", &JSONFormatter{})
	require.NoError(t, err)
	err = b.RegisterNode("n1", &FileSink{})
	require.NoError(t, err)
}

// TestBroker_RegisterNode_AllowOverwrite_Explicit is used to prove that nodes can be
// overwritten when a Broker has been explicitly configured with the AllowOverwrite policy.
func TestBroker_RegisterNode_AllowOverwrite_Explicit(t *testing.T) {
	b, err := NewBroker(WithNodeRegistrationPolicy(AllowOverwrite))
	require.NoError(t, err)
	err = b.RegisterNode("n1", &JSONFormatter{})
	require.NoError(t, err)
	err = b.RegisterNode("n1", &FileSink{})
	require.NoError(t, err)
}

// TestBroker_RegisterNode_DenyOverwrite is used to prove that nodes can't be
// overwritten when a Broker has been configured with the DenyOverwrite policy.
func TestBroker_RegisterNode_DenyOverwrite(t *testing.T) {
	b, err := NewBroker(WithNodeRegistrationPolicy(DenyOverwrite))
	require.NoError(t, err)
	err = b.RegisterNode("n1", &JSONFormatter{})
	require.NoError(t, err)
	err = b.RegisterNode("n1", &FileSink{})
	require.Error(t, err)
	require.EqualError(t, err, "node ID \"n1\" is already registered, configured policy prevents overwriting")
}

// TestBroker_RegisterPipeline_AllowOverwrite_Implicit is used to prove that pipelines can be
// overwritten when a Broker has been implicitly configured with the AllowOverwrite policy.
// This is the default in order to maintain pre-existing behavior.
func TestBroker_RegisterPipeline_AllowOverwrite_Implicit(t *testing.T) {
	b, err := NewBroker()
	require.NoError(t, err)

	err = b.RegisterNode("f1", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("f2", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("s1", &FileSink{})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f1", "s1"},
	})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f2", "s1"},
	})
	require.NoError(t, err)
}

// TestBroker_RegisterPipeline_AllowOverwrite_Explicit is used to prove that pipelines can be
// overwritten when a Broker has been explicitly configured with the AllowOverwrite policy.
func TestBroker_RegisterPipeline_AllowOverwrite_Explicit(t *testing.T) {
	b, err := NewBroker(WithPipelineRegistrationPolicy(AllowOverwrite))
	require.NoError(t, err)

	err = b.RegisterNode("f1", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("f2", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("s1", &FileSink{})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f1", "s1"},
	})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f2", "s1"},
	})
	require.NoError(t, err)
}

// TestBroker_RegisterPipeline_DenyOverwrite is used to prove that pipelines can't
// // be overwritten when a Broker has been configured with the DenyOverwrite policy.
func TestBroker_RegisterPipeline_DenyOverwrite(t *testing.T) {
	b, err := NewBroker(WithPipelineRegistrationPolicy(DenyOverwrite))
	require.NoError(t, err)

	err = b.RegisterNode("f1", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("f2", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterNode("s1", &FileSink{})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f1", "s1"},
	})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f2", "s1"},
	})
	require.Error(t, err)
	require.EqualError(t, err, "pipeline ID \"p1\" is already registered, configured policy prevents overwriting")
}

func TestBroker_RegisterPipeline_WithCloser(t *testing.T) {
	b, err := NewBroker(WithPipelineRegistrationPolicy(DenyOverwrite))
	require.NoError(t, err)

	mc := &mockCloserWithWrapper{n: &mockCloser{}}
	err = b.RegisterNode("mc1", mc)
	require.NoError(t, err)

	err = b.RegisterNode("f1", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f1", "mc1"},
	})
	require.NoError(t, err)

	ok, err := b.RemovePipelineAndNodes("t", "p1")
	require.NoError(t, err)
	require.True(t, ok)

	assert.True(t, mc.n.closed)
}

func TestBroker_RegisterPipeline_WithCloserError(t *testing.T) {
	b, err := NewBroker(WithPipelineRegistrationPolicy(DenyOverwrite))
	require.NoError(t, err)

	mc := &mockCloserWithWrapper{n: &mockCloser{closeErr: errors.New("close error")}}
	err = b.RegisterNode("mc1", mc)
	require.NoError(t, err)

	err = b.RegisterNode("f1", &JSONFormatter{})
	require.NoError(t, err)

	err = b.RegisterPipeline(Pipeline{
		PipelineID: "p1",
		EventType:  "t",
		NodeIDs:    []NodeID{"f1", "mc1"},
	})
	require.NoError(t, err)

	ok, err := b.RemovePipelineAndNodes("t", "p1")
	require.Error(t, err)
	require.True(t, ok)

	assert.Contains(t, err.Error(), "close error")
	assert.False(t, mc.n.closed)
}
