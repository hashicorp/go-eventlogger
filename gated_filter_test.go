package eventlogger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
		name             string
		gf               *eventlogger.GatedFilter
		setupEvents      []*eventlogger.Event
		ignoreTimestamps bool
		testEvent        *eventlogger.Event
		wantEvent        *eventlogger.Event
		wantErr          bool
		wantErrContains  string
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
					Details []eventlogger.SimpleGatedDetailsPayload
				}{
					ID: "event-1",
					Header: map[string]interface{}{
						"roles": []string{"admin", "anon"},
						"tmz":   "EST",
						"user":  "alice",
					},
					Details: []eventlogger.SimpleGatedDetailsPayload{
						{
							Type:      "test",
							CreatedAt: now.String(),
							Payload: map[string]interface{}{
								"file_name":   "file1.txt",
								"total_bytes": 1024,
							},
						},
						{
							Type:      "test",
							CreatedAt: now.String(),
							Payload: map[string]interface{}{
								"file_name":   "file2.txt",
								"total_bytes": 512,
							},
						},
						{
							Type:      "test",
							CreatedAt: now.String(),
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
			name: "expired-no-broker",
			gf: &eventlogger.GatedFilter{
				Expiration: 1 * time.Nanosecond,
			},
			ignoreTimestamps: true,
			setupEvents:      setupEvents,
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
				Type: "test",
				// not setting CreatedAt because ignoreTimestamps == true
				Payload: struct {
					ID      string
					Header  map[string]interface{}
					Details []eventlogger.SimpleGatedDetailsPayload
				}{
					ID: "event-1",
					Details: []eventlogger.SimpleGatedDetailsPayload{
						{
							Type:      "test",
							CreatedAt: now.String(),
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
			if tt.ignoreTimestamps {
				tt.wantEvent.CreatedAt = got.CreatedAt
			}
			assert.Equal(tt.wantEvent, got)
		})
	}
	t.Run("expiration-with-broker", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		b, cleanup, tmpDir := testBroker(t, "expiration-with-broker", "test")
		defer cleanup()

		gf := &eventlogger.GatedFilter{
			Expiration: 1 * time.Nanosecond,
			Broker:     b,
		}
		gated, err := gf.Process(ctx, setupEvents[0])
		require.NoError(err)
		require.Empty(gated)

		got, err := gf.Process(ctx, &eventlogger.Event{
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
		})
		require.NoError(err)
		wantEvent := &eventlogger.Event{
			Type:      "test",
			CreatedAt: got.CreatedAt,
			Payload: struct {
				ID      string
				Header  map[string]interface{}
				Details []eventlogger.SimpleGatedDetailsPayload
			}{
				ID: "event-1",
				Details: []eventlogger.SimpleGatedDetailsPayload{
					{
						Type:      "test",
						CreatedAt: now.String(),
						Payload: map[string]interface{}{
							"file_name":   "file3.txt",
							"total_bytes": 1000000,
						},
					},
				},
			},
		}
		assert.Equal(wantEvent, got)

		// Check the contents of the log
		files, err := ioutil.ReadDir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) > 1 {
			t.Errorf("Expected 1 log file, got %d", len(files))
		}

		dat, err := ioutil.ReadFile(filepath.Join(tmpDir, files[0].Name()))
		if err != nil {
			t.Fatal(err)
		}

		type loggedEvent struct {
			CreatedAt string `json:"created_at"`
			EventType string `json:"event_type"`
			Payload   struct {
				ID      string                 `json:"id"`
				Header  map[string]interface{} `json:"header,omitempty"`
				Details []struct {
					Type      string                 `json:"type"`
					CreatedAt string                 `json:"created_at"`
					Payload   map[string]interface{} `json:"payload"`
				}
			}
		}
		gotEvent := &loggedEvent{}
		require.NoError(json.Unmarshal(dat, gotEvent))

		wantReadEvent := &loggedEvent{
			CreatedAt: gotEvent.CreatedAt,
			EventType: "test",
			Payload: struct {
				ID      string                 "json:\"id\""
				Header  map[string]interface{} "json:\"header,omitempty\""
				Details []struct {
					Type      string                 "json:\"type\""
					CreatedAt string                 "json:\"created_at\""
					Payload   map[string]interface{} "json:\"payload\""
				}
			}{
				ID: "event-1",
				Header: map[string]interface{}{
					"tmz":  "EST",
					"user": "alice",
				},
				Details: []struct {
					Type      string                 "json:\"type\""
					CreatedAt string                 "json:\"created_at\""
					Payload   map[string]interface{} "json:\"payload\""
				}{
					{
						Type:      "test",
						CreatedAt: now.String(),
						Payload: map[string]interface{}{
							"file_name":   "file1.txt",
							"total_bytes": float64(1024),
						},
					},
				},
			},
		}
		assert.Equal(wantReadEvent, gotEvent)

	})

}

func testBroker(t *testing.T, testName string, eventType string) (*eventlogger.Broker, func(), string) {
	t.Helper()
	require := require.New(t)
	require.NotEmpty(eventType)
	tmpDir, err := ioutil.TempDir("", testName)
	require.NoError(err)
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Marshal to JSON
	n1 := &eventlogger.JSONFormatter{}
	// Send to FileSink
	n2 := &eventlogger.FileSink{Path: tmpDir, FileName: "file.log"}

	// Create a broker
	b := eventlogger.NewBroker()

	// Register the graph with the broker
	et := eventlogger.EventType(eventType)
	nodes := []eventlogger.Node{n1, n2}
	nodeIDs := make([]eventlogger.NodeID, len(nodes))
	for i, node := range nodes {
		id := eventlogger.NodeID(fmt.Sprintf("node-%d", i))
		err := b.RegisterNode(id, node)
		if err != nil {
			t.Fatal(err)
		}
		nodeIDs[i] = id
	}
	err = b.RegisterPipeline(eventlogger.Pipeline{
		EventType:  et,
		PipelineID: "id",
		NodeIDs:    nodeIDs,
	})
	require.NoError(err)
	return b, cleanup, tmpDir
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
