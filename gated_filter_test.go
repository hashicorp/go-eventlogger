package eventlogger_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatedFilter_Process(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()
	testGf := &eventlogger.GatedFilter{
		NowFunc: func() time.Time { return now },
	}

	setupEvents := []*eventlogger.Event{
		{
			Type:      "test",
			CreatedAt: now,
			Payload: &eventlogger.SimpleGatedPayload{
				ID: "event-1",
				Header: map[string]interface{}{
					"user": "alice",
					"tmz":  "EST",
				},
				Detail: map[string]interface{}{
					"file_name":   "file1.txt",
					"total_bytes": 1024,
				},
			},
		},
		{
			Type:      "test",
			CreatedAt: now,
			Payload: &eventlogger.SimpleGatedPayload{
				ID: "event-1",
				Header: map[string]interface{}{
					"roles": []string{"admin", "anon"},
				},
				Detail: map[string]interface{}{
					"file_name":   "file2.txt",
					"total_bytes": 512,
				},
			},
		},
	}

	tests := []struct {
		name            string
		gf              *eventlogger.GatedFilter
		setupEvents     []*eventlogger.Event
		testEvent       *eventlogger.Event
		wantEvent       *eventlogger.Event
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "simple",
			gf:          testGf,
			setupEvents: setupEvents,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &eventlogger.SimpleGatedPayload{
					ID:    "event-1",
					Flush: true,
					Detail: map[string]interface{}{
						"file_name":   "file3.txt",
						"total_bytes": 1000000,
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: struct {
					ID      string
					Header  map[string]interface{}
					Details []*eventlogger.Event
				}{
					ID: "event-1",
					Header: map[string]interface{}{
						"roles": []string{"admin", "anon"},
						"tmz":   "EST",
						"user":  "alice",
					},
					Details: []*eventlogger.Event{
						{
							Type:      "test",
							CreatedAt: now,
							Payload: map[string]interface{}{
								"file_name":   "file1.txt",
								"total_bytes": 1024,
							},
						},
						{
							Type:      "test",
							CreatedAt: now,
							Payload: map[string]interface{}{
								"file_name":   "file2.txt",
								"total_bytes": 512,
							},
						},
						{
							Type:      "test",
							CreatedAt: now,
							Payload: map[string]interface{}{
								"file_name":   "file3.txt",
								"total_bytes": 1000000,
							},
						},
					},
				},
			},
		},
		{
			name:            "missing-event",
			gf:              testGf,
			wantErr:         true,
			wantErrContains: "missing event",
		},
		{
			name: "not-gateable",
			gf:   testGf,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "not-gateable",
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "not-gateable",
			},
		},
		{
			name: "missing-id",
			gf:   testGf,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &eventlogger.SimpleGatedPayload{
					Header: map[string]interface{}{
						"missing-id": true,
					},
				},
			},
			wantErr:         true,
			wantErrContains: "missing ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			for _, e := range tt.setupEvents {
				_, err := tt.gf.Process(ctx, e)
				require.NoError(err)
			}
			got, err := tt.gf.Process(ctx, tt.testEvent)
			if tt.wantErr {
				require.Error(err)
				assert.Nil(got)
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantEvent, got)
		})
	}

}

func TestGatedFilter_Now(t *testing.T) {
	t.Parallel()
	t.Run("default-now", func(t *testing.T) {
		assert := assert.New(t)
		gf := eventlogger.GatedFilter{}
		n := time.Now()
		got := gf.Now()
		assert.True(got.Before(time.Now()))
		assert.True(got.After(n))
	})
	t.Run("override-now", func(t *testing.T) {
		assert := assert.New(t)
		n := time.Now()
		gf := eventlogger.GatedFilter{
			NowFunc: func() time.Time { return n },
		}
		assert.Equal(n, gf.Now())
	})
}

func TestGatedFilter_Type(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	gf := eventlogger.GatedFilter{}
	assert.Equal(eventlogger.NodeTypeFilter, gf.Type())
}
