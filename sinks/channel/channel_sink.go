// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/eventlogger"
)

// ChannelSink is a sink node which sends
// the event to a channel
type ChannelSink struct {
	eventChan chan<- *eventlogger.Event

	// The time to wait for a write before returning an error
	timeoutDuration time.Duration
}

var _ eventlogger.Node = &ChannelSink{}

// newChannelSink creates a ChannelSink
// The time.Duration value is used to set a timeout on the consumer for sending events
// This is to account for consumers having different timeouts than senders
func NewChannelSink(c chan<- *eventlogger.Event, t time.Duration) (*ChannelSink, error) {
	if c == nil {
		return nil, errors.New("missing event channel")
	}
	if t <= 0 {
		return nil, errors.New("duration must be greater than 0")
	}

	return &ChannelSink{
		eventChan:       c,
		timeoutDuration: t,
	}, nil
}

// Process sends the event on a channel
// Process will wait for the ChannelSink timeoutDuration for a write before returning an error
// This is to account for consumers having different timeouts than senders
// Returns a nil event as this is a leaf node
func (c *ChannelSink) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	select {
	case c.eventChan <- e:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(c.timeoutDuration):
		return nil, fmt.Errorf("chan write timeout after %s", c.timeoutDuration)
	}

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
