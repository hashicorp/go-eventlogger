// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package channel

import (
	"testing"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChannelSink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		c               chan *eventlogger.Event
		wantErrContains string
	}{
		{
			name:            "missing-channel",
			wantErrContains: "missing event channel",
		},
		{
			name: "valid",
			c:    make(chan *eventlogger.Event),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			got, err := NewChannelSink(tt.c)
			if tt.wantErrContains != "" {
				require.Error(err)
				assert.Contains(err.Error(), tt.wantErrContains)
				return
			}
			assert.NotNil(got)
		})
	}
}
