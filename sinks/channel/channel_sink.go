// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel

import (
	"context"
	"errors"
	"sync"

	"github.com/hashicorp/eventlogger"
)

// ChannelSink is a sink node which sends
// the event to a channel
type ChannelSink struct {
	mu sync.Mutex

	eventChan chan *eventlogger.Event
}

var _ eventlogger.Node = &ChannelSink{}

// newChannelSink creates a ChannelSink
func NewChannelSink(c chan *eventlogger.Event) (*ChannelSink, error) {
	if c == nil {
		return nil, errors.New("missing event channel")
	}

	return &ChannelSink{
		eventChan: c,
	}, nil
}

// Process sends the event on a channel
// Returns a nil event as this is a leaf node
func (c *ChannelSink) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventChan <- e

	return nil, nil
}

// Reopen is a no op
func (c *ChannelSink) Reopen() error {
	return nil
}

// Type describes the type of the node as a NodeTypeSink.
func (c *ChannelSink) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeSink
}

// Name returns a representation of the ChannelSink's name
func (c *ChannelSink) Name() string {
	return "ChannelSink"
}
