// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Broker is the top-level entity used in the library for configuring the system
// and for sending events.
//
// Brokers have registered Nodes which may be composed into registered Pipelines
// for EventTypes.
//
// A Node may be a filter, formatter or sink (see NodeType).
//
// A Broker may have multiple Pipelines.
//
// EventTypes may have multiple Pipelines.
//
// A Pipeline for an EventType may contain multiple filters, one formatter and
// one sink.
//
// If a Pipeline does not have a formatter, then the event will not be written
// to the Sink.
//
// A Node can be shared across multiple pipelines.
type Broker struct {
	nodes  map[NodeID]*nodeUsage
	graphs map[EventType]*graph
	lock   sync.RWMutex

	*clock
}

// nodeUsage tracks how many times a Node is referenced by registered pipelines.
type nodeUsage struct {
	node       Node
	references int
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		nodes:  make(map[NodeID]*nodeUsage),
		graphs: make(map[EventType]*graph),
	}
}

// clock only exists to make testing simpler.
type clock struct {
	now time.Time
}

// Now returns the current time
func (c *clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.now
}

// StopTimeAt allows you to "stop" the Broker's timestamp clock at a predicable
// point in time, so timestamps are predictable for testing.
func (b *Broker) StopTimeAt(now time.Time) {
	b.clock = &clock{now: now}
}

// Status describes the result of a Send.
type Status struct {
	// complete lists the IDs of 'filter' and 'sink' type nodes that successfully processed the Event.
	complete []NodeID
	// complete lists the IDs of 'sink' type nodes that successfully processed the Event.
	completeSinks []NodeID
	// Warnings lists any non-fatal errors that occurred while sending an Event.
	Warnings []error
}

func (s Status) getError(threshold, thresholdSinks int) error {
	switch {
	case len(s.complete) < threshold:
		return fmt.Errorf("event not processed by enough 'filter' and 'sink' nodes")
	case len(s.completeSinks) < thresholdSinks:
		return fmt.Errorf("event not processed by enough 'sink' nodes")
	default:
		return nil
	}
}

// Send writes an event of type t to all registered pipelines concurrently and
// reports on the result.  An error will only be returned if a pipeline's delivery
// policies could not be satisfied.
func (b *Broker) Send(ctx context.Context, t EventType, payload interface{}) (Status, error) {
	b.lock.RLock()
	g, ok := b.graphs[t]
	b.lock.RUnlock()

	if !ok {
		return Status{}, fmt.Errorf("no graph for EventType %s", t)
	}

	e := &Event{
		Type:      t,
		CreatedAt: b.clock.Now(),
		Formatted: make(map[string][]byte),
		Payload:   payload,
	}

	return g.process(ctx, e)
}

// Reopen calls every registered Node's Reopen() function.  The intention is to
// ask all nodes to reopen any files they have open.  This is typically used as
// part of log rotation: after rotating, the rotator sends a signal to the
// application, which then would invoke this method.  Another typically use-case
// is to have all Nodes reevaluated any external configuration they might have.
func (b *Broker) Reopen(ctx context.Context) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, g := range b.graphs {
		if err := g.reopen(ctx); err != nil {
			return err
		}
	}

	return nil
}

// NodeID is a string that uniquely identifies a Node.
type NodeID string

// RegisterNode assigns a node ID to a node.  Node IDs should be unique. A Node
// may be a filter, formatter or sink (see NodeType). Nodes can be shared across
// multiple pipelines.
func (b *Broker) RegisterNode(id NodeID, node Node) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	nr := &nodeUsage{node: node, references: 0}

	// Check if this node is already registered, if so maintain reference count
	r, exists := b.nodes[id]
	if exists {
		nr.references = r.references
	}

	b.nodes[id] = nr

	return nil
}

// PipelineID is a string that uniquely identifies a Pipeline within a given EventType.
type PipelineID string

// Pipeline defines a pipe: its ID, the EventType it's for, and the nodes
// that it contains. Nodes can be shared across multiple pipelines.
type Pipeline struct {
	// PipelineID uniquely identifies the Pipeline
	PipelineID PipelineID

	// EventType defines the type of event the Pipeline processes
	EventType EventType

	// NodeIDs defines Pipeline's the list of nodes
	NodeIDs []NodeID
}

// RegisterPipeline adds a pipeline to the broker.
func (b *Broker) RegisterPipeline(def Pipeline) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[def.EventType]
	if !ok {
		g = &graph{}
		b.graphs[def.EventType] = g
	}

	// Gather the registered nodes, so they can be referenced for this pipeline.
	nodes := make([]Node, len(def.NodeIDs))
	for i, n := range def.NodeIDs {
		nodeUsage, ok := b.nodes[n]
		if !ok {
			return fmt.Errorf("nodeID %q not registered", n)
		}
		nodes[i] = nodeUsage.node
	}

	root, err := linkNodes(nodes, def.NodeIDs)
	if err != nil {
		return err
	}

	err = g.doValidate(nil, root)
	if err != nil {
		return err
	}

	// Store the pipeline and then update the reference count of the nodes in that pipeline.
	g.roots.Store(def.PipelineID, root)
	for _, id := range def.NodeIDs {
		nodeUsage, ok := b.nodes[id]
		// We can be optimistic about this as we would have already errored above.
		if ok {
			nodeUsage.references++
		}
	}

	return nil
}

// RemovePipeline removes a pipeline from the broker.
func (b *Broker) RemovePipeline(t EventType, id PipelineID) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		return fmt.Errorf("no graph for EventType %s", t)
	}

	g.roots.Delete(id)
	return nil
}

// RemovePipelineAndNodes will attempt to remove all nodes referenced by the pipeline.
// Any nodes that are referenced by other pipelines will not be removed.
func (b *Broker) RemovePipelineAndNodes(t EventType, id PipelineID) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		return fmt.Errorf("no graph for EventType %s", t)
	}

	nodes, err := g.roots.Nodes(id)
	if err != nil {
		return fmt.Errorf("unable to retrieve all nodes referenced by pipeline ID %q: %w", id, err)
	}

	g.roots.Delete(id)

	for _, nodeID := range nodes {
		nodeUsage, ok := b.nodes[nodeID]
		if !ok {
			return fmt.Errorf("pipeline ID %q: node ID %q is not registered", id, nodeID)
		}

		switch nodeUsage.references {
		case 0, 1:
			// Node is not currently in use, or was only being used by this pipeline
			delete(b.nodes, nodeID)
		default:
			nodeUsage.references--
		}
	}

	return nil
}

// SetSuccessThreshold sets the success threshold per eventType.  For the
// overall processing of a given event to be considered a success, at least as
// many pipelines as the threshold value must successfully process the event.
// This means that a filter could of course filter an event before it reaches
// the pipeline's sink, but it would still count as success when it comes to
// meeting this threshold.  Use this when you want to allow the filtering of
// events without causing an error because an event was filtered.
func (b *Broker) SetSuccessThreshold(t EventType, successThreshold int) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if successThreshold < 0 {
		return fmt.Errorf("successThreshold must be 0 or greater")
	}

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{}
		b.graphs[t] = g
	}

	g.successThreshold = successThreshold
	return nil
}

// SetSuccessThresholdSinks sets the success threshold per eventType.  For the
// overall processing of a given event to be considered a success, at least as
// many sinks as the threshold value must successfully process the event.
func (b *Broker) SetSuccessThresholdSinks(t EventType, successThresholdSinks int) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if successThresholdSinks < 0 {
		return fmt.Errorf("successThresholdSinks must be 0 or greater")
	}

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{}
		b.graphs[t] = g
	}

	g.successThresholdSinks = successThresholdSinks
	return nil
}
