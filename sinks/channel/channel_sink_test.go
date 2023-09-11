// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel

import (
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
