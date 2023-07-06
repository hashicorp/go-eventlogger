// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
)

// RegistrationPolicy is used to specify what kind of policy should apply when
// registering components (e.g. Pipeline, Node) with the Broker
type RegistrationPolicy string

const (
	AllowOverwrite RegistrationPolicy = "AllowOverwrite"
	DenyOverwrite  RegistrationPolicy = "DenyOverwrite"
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

	pipelineRegistrationPolicy RegistrationPolicy
	nodeRegistrationPolicy     RegistrationPolicy

	*clock
}

// nodeUsage tracks how many times a Node is referenced by registered pipelines.
type nodeUsage struct {
	node           Node
	referenceCount int
}

// Option allows options to be passed as arguments.
type Option func(*options) error

// options are used to represent configuration for the broker.
type options struct {
	withPipelineRegistrationPolicy RegistrationPolicy
	withNodeRegistrationPolicy     RegistrationPolicy
}

// getDefaultOptions returns a set of default options
func getDefaultOptions() options {
	return options{
		withPipelineRegistrationPolicy: AllowOverwrite,
		withNodeRegistrationPolicy:     AllowOverwrite,
	}
}

// getOpts iterates the inbound Options and returns a struct.
// Each Option is applied in the order it appears in the argument list, so it is
// possible to supply the same Option numerous times and the 'last write wins'.
func getOpts(opt ...Option) (options, error) {
	opts := getDefaultOptions()
	for _, o := range opt {
		if o == nil {
			continue
		}
		if err := o(&opts); err != nil {
			return options{}, err
		}
	}
	return opts, nil
}

// WithPipelineRegistrationPolicy configures the option that determines the pipeline registration policy.
func WithPipelineRegistrationPolicy(policy RegistrationPolicy) Option {
	return func(o *options) error {
		var err error

		switch policy {
		case AllowOverwrite, DenyOverwrite:
			o.withPipelineRegistrationPolicy = policy
		default:
			err = fmt.Errorf("'%s' is not a valid pipeline registration policy: %w", policy, ErrInvalidParameter)
		}

		return err
	}
}

// WithNodeRegistrationPolicy configures the option that determines the node registration policy.
func WithNodeRegistrationPolicy(policy RegistrationPolicy) Option {
	return func(o *options) error {
		var err error

		switch policy {
		case AllowOverwrite, DenyOverwrite:
			o.withNodeRegistrationPolicy = policy
		default:
			err = fmt.Errorf("'%s' is not a valid node registration policy: %w", policy, ErrInvalidParameter)
		}

		return err
	}
}

// NewBroker creates a new Broker applying any supplied options.
func NewBroker(opt ...Option) (*Broker, error) {
	opts, err := getOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("cannot create broker: %w", err)
	}

	b := &Broker{
		nodes:  make(map[NodeID]*nodeUsage),
		graphs: make(map[EventType]*graph),
	}
	if opts.withPipelineRegistrationPolicy != "" {
		b.pipelineRegistrationPolicy = opts.withPipelineRegistrationPolicy
	}
	if opts.withNodeRegistrationPolicy != "" {
		b.nodeRegistrationPolicy = opts.withNodeRegistrationPolicy
	}

	return b, nil
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
	if id == "" {
		return errors.New("unable to register node, node ID cannot be empty")
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	nr := &nodeUsage{node: node, referenceCount: 0}

	// Check if this node is already registered, if so maintain reference count
	r, exists := b.nodes[id]
	if exists {
		switch b.nodeRegistrationPolicy {
		case AllowOverwrite:
			nr.referenceCount = r.referenceCount
		case DenyOverwrite:
			return fmt.Errorf("node ID %q is already registered, configured policy prevents overwriting", id)
		}
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
	err := def.validate()
	if err != nil {
		return err
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	g, exists := b.graphs[def.EventType]
	if !exists {
		g = &graph{}
		b.graphs[def.EventType] = g
	} else if b.pipelineRegistrationPolicy == DenyOverwrite {
		return fmt.Errorf("pipeline ID %q is already registered, configured policy prevents overwriting", def.PipelineID)
	}

	// Gather the registered nodes, so they can be referenced for this pipeline.
	nodes := make([]Node, len(def.NodeIDs))
	for i, n := range def.NodeIDs {
		nodeUsage, ok := b.nodes[n]
		if !ok {
			return fmt.Errorf("node ID %q not registered", n)
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
			nodeUsage.referenceCount++
		}
	}

	return nil
}

// RemovePipeline removes a pipeline from the broker.
func (b *Broker) RemovePipeline(t EventType, id PipelineID) error {
	switch {
	case t == "":
		return errors.New("event type cannot be empty")
	case id == "":
		return errors.New("pipeline ID cannot be empty")
	}

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
//
// Failed preconditions will result in a return of false with an error and
// neither the pipeline nor nodes will be deleted.
//
// Once we start deleting the pipeline and nodes, we will continue until completion,
// but we'll return true along with any errors encountered (as multierror.Error).
func (b *Broker) RemovePipelineAndNodes(ctx context.Context, t EventType, id PipelineID) (bool, error) {
	switch {
	case t == "":
		return false, errors.New("event type cannot be empty")
	case id == "":
		return false, errors.New("pipeline ID cannot be empty")
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		return false, fmt.Errorf("no graph for EventType %s", t)
	}

	nodes, err := g.roots.Nodes(id)
	if err != nil {
		return false, fmt.Errorf("unable to retrieve all nodes referenced by pipeline ID %q: %w", id, err)
	}

	g.roots.Delete(id)

	var nodeErr error

	for _, nodeID := range nodes {
		nodeUsage, ok := b.nodes[nodeID]
		if !ok {
			// We might get multiple nodes which cannot be found
			nodeErr = multierror.Append(nodeErr, fmt.Errorf("node not found: %q", nodeID))
			continue
		}

		switch nodeUsage.referenceCount {
		case 0, 1:
			nc := NewNodeController(nodeUsage.node)
			if err := nc.Close(ctx); err != nil {
				nodeErr = multierror.Append(nodeErr, fmt.Errorf("unable to close node ID %q: %w", nodeID, err))
			}
			// Node is not currently in use, or was only being used by this pipeline
			delete(b.nodes, nodeID)
		default:
			nodeUsage.referenceCount--
		}
	}

	return true, nodeErr
}

// SetSuccessThreshold sets the success threshold per eventType.  For the
// overall processing of a given event to be considered a success, at least as
// many pipelines as the threshold value must successfully process the event.
// This means that a filter could of course filter an event before it reaches
// the pipeline's sink, but it would still count as success when it comes to
// meeting this threshold.  Use this when you want to allow the filtering of
// events without causing an error because an event was filtered.
func (b *Broker) SetSuccessThreshold(t EventType, successThreshold int) error {
	switch {
	case t == "":
		return errors.New("event type cannot be empty")
	case successThreshold < 0:
		return fmt.Errorf("successThreshold must be 0 or greater")
	}

	b.lock.Lock()
	defer b.lock.Unlock()

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
	switch {
	case t == "":
		return errors.New("event type cannot be empty")
	case successThresholdSinks < 0:
		return fmt.Errorf("successThresholdSinks must be 0 or greater")
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	g, ok := b.graphs[t]
	if !ok {
		g = &graph{}
		b.graphs[t] = g
	}

	g.successThresholdSinks = successThresholdSinks
	return nil
}

// validate ensures that the Pipeline has the required configuration to allow
// registration, removal or usage, without issue.
func (p Pipeline) validate() error {
	var err error

	if p.PipelineID == "" {
		err = multierror.Append(err, errors.New("pipeline ID is required"))
	}

	if p.EventType == "" {
		err = multierror.Append(err, errors.New("event type is required"))
	}

	if len(p.NodeIDs) == 0 {
		err = multierror.Append(err, errors.New("node IDs are required"))
	}

	for _, n := range p.NodeIDs {
		if n == "" {
			err = multierror.Append(err, errors.New("node ID cannot be empty"))
			break
		}
	}

	return err
}
