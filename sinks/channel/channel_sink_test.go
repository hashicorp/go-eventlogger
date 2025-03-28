// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChannelSink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		c               chan *eventlogger.Event
		d               time.Duration
		wantErrContains string
	}{
		{
			name:            "missing-channel",
			d:               time.Second,
			wantErrContains: "missing event channel",
		},
		{
			name:            "missing-duration",
			c:               make(chan *eventlogger.Event),
			wantErrContains: "duration must be greater than 0",
		},
		{
			name: "valid",
			c:    make(chan *eventlogger.Event),
			d:    time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			got, err := NewChannelSink(tt.c, tt.d)
			if tt.wantErrContains != "" {
				require.Error(err)
				assert.Contains(err.Error(), tt.wantErrContains)
				return
			}
			assert.NotNil(got)
		})
	}
}

func TestProcess(t *testing.T) {
	t.Parallel()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	tests := []struct {
		name            string
		c               chan *eventlogger.Event
		ctx             context.Context
		d               time.Duration
		wantErrContains string
	}{
		{
			name: "valid",
			c:    make(chan *eventlogger.Event, 1),
			d:    time.Second,
			ctx:  context.Background(),
		},
		{
			name:            "write timeout",
			c:               make(chan *eventlogger.Event),
			d:               time.Second,
			ctx:             context.Background(),
			wantErrContains: "chan write timeout",
		},
		{
			name:            "context timeout",
			c:               make(chan *eventlogger.Event),
			d:               time.Second,
			ctx:             timeoutCtx,
			wantErrContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			sink, err := NewChannelSink(tt.c, tt.d)
			require.Nil(err)
			assert.NotNil(sink)
			event := &eventlogger.Event{
				Type:      "testEvent",
				CreatedAt: time.Time{},
				Formatted: nil,
				Payload:   nil,
			}
			got, err := sink.Process(tt.ctx, event)
			require.Nil(got)
			if tt.wantErrContains != "" {
				require.Error(err)
				assert.Contains(err.Error(), tt.wantErrContains)
				return
			}
			require.Nil(err)
		})
	}
}
