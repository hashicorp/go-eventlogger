package gated_test

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
	"github.com/hashicorp/eventlogger/filters/gated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatedFilter_Process(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()
	testGf := &gated.Filter{
		NowFunc: func() time.Time { return now },
	}

	setupEvents := []*eventlogger.Event{
		{
			Type:      "test",
			CreatedAt: now,
			Payload: &gated.Payload{
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
			Payload: &gated.Payload{
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
		gf               *gated.Filter
		setupEvents      []*eventlogger.Event
		ignoreTimestamps bool
		testEvent        *eventlogger.Event
		wantEvent        *eventlogger.Event
		wantErr          bool
		wantErrContains  string
		wantErrIs        error
	}{
		{
			name:        "simple",
			gf:          testGf,
			setupEvents: setupEvents,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &gated.Payload{
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
				Formatted: map[string][]byte{},
				Payload: gated.EventPayload{
					ID: "event-1",
					Header: map[string]interface{}{
						"roles": []string{"admin", "anon"},
						"tmz":   "EST",
						"user":  "alice",
					},
					Details: []gated.EventPayloadDetails{
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
			gf: &gated.Filter{
				Expiration: 1 * time.Nanosecond,
			},
			ignoreTimestamps: true,
			setupEvents:      setupEvents,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &gated.Payload{
					ID:    "event-1",
					Flush: true,
					Detail: map[string]interface{}{
						"file_name":   "file3.txt",
						"total_bytes": 1000000,
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Formatted: map[string][]byte{},
				Type:      "test",
				// not setting CreatedAt because ignoreTimestamps == true
				Payload: gated.EventPayload{
					ID: "event-1",
					Details: []gated.EventPayloadDetails{
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
			wantErrIs:       eventlogger.ErrInvalidParameter,
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
				Payload: &gated.Payload{
					Header: map[string]interface{}{
						"missing-id": true,
					},
				},
			},
			wantErr:         true,
			wantErrContains: "missing ID",
			wantErrIs:       eventlogger.ErrInvalidParameter,
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
				if tt.wantErrIs != nil {
					assert.ErrorIs(err, eventlogger.ErrInvalidParameter)
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
		_, gf, tmpDir := testBrokerWithGatedFilter(t, "expiration-with-broker", "test")

		gf.Expiration = 1 * time.Nanosecond

		got, err := gf.Process(ctx, setupEvents[0])
		require.NoError(err)
		require.Empty(got)

		got, err = gf.Process(ctx, &eventlogger.Event{
			Type:      "test",
			CreatedAt: now,
			Payload: &gated.Payload{
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
			Formatted: map[string][]byte{},
			Payload: gated.EventPayload{
				ID: "event-1",
				Details: []gated.EventPayloadDetails{
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
		require.NoError(err)
		if len(files) > 1 {
			t.Errorf("Expected 1 log file, got %d", len(files))
		}

		dat, err := ioutil.ReadFile(filepath.Join(tmpDir, files[0].Name()))
		require.NoError(err)

		type loggedEvent struct {
			CreatedAt string `json:"created_at"`
			EventType string `json:"event_type"`
			Payload   gated.EventPayload
		}
		gotEvent := &loggedEvent{}
		require.NoError(json.Unmarshal(dat, gotEvent))

		wantReadEvent := &loggedEvent{
			CreatedAt: gotEvent.CreatedAt,
			EventType: "test",
			Payload: gated.EventPayload{
				ID: "event-1",
				Header: map[string]interface{}{
					"tmz":  "EST",
					"user": "alice",
				},
				Details: []gated.EventPayloadDetails{
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

func TestGatedFilter_FlushAll(t *testing.T) {
	t.Parallel()
	now := time.Now()

	type loggedEvent struct {
		CreatedAt string `json:"created_at"`
		EventType string `json:"event_type"`
		Payload   gated.EventPayload
	}

	tests := []struct {
		name      string
		t         eventlogger.EventType
		payload   *gated.Payload
		wantEvent *loggedEvent
		wantErr   bool
	}{
		{
			name: "success",
			t:    eventlogger.EventType("test"),
			payload: &gated.Payload{
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
			wantEvent: &loggedEvent{
				EventType: "test",
				Payload: gated.EventPayload{
					ID: "event-1",
					Header: map[string]interface{}{
						"tmz":  "EST",
						"user": "alice",
					},
					Details: []gated.EventPayloadDetails{
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
			},
		},
		{
			name: "no-gated-events",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			ctx := context.Background()
			b, gf, tmpDir := testBrokerWithGatedFilter(t, tt.name, "test")
			gf.Expiration = 100 * time.Minute // Note: be very careful setting the exp to something so large
			gf.NowFunc = func() time.Time { return now }

			if tt.payload != nil {
				_, err := b.Send(ctx, tt.t, tt.payload)
				require.NoError(err)
			}

			err := gf.FlushAll(ctx)
			if tt.wantErr {
				require.Error(err)
				return
			}
			// Check the contents of the log
			files, err := ioutil.ReadDir(tmpDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(files) > 1 {
				t.Errorf("Expected 1 log file, got %d", len(files))
			}
			switch tt.wantEvent == nil {
			case true:
				assert.Len(files, 0)
			default:
				require.Len(files, 1)
				dat, err := ioutil.ReadFile(filepath.Join(tmpDir, files[0].Name()))
				require.NoError(err)

				gotEvent := &loggedEvent{}
				require.NoError(json.Unmarshal(dat, gotEvent))
				tt.wantEvent.CreatedAt = gotEvent.CreatedAt
				tt.wantEvent.Payload.Details[0].CreatedAt = gotEvent.Payload.Details[0].CreatedAt
				assert.Equal(tt.wantEvent, gotEvent)
			}
		})
	}
	name := "no-broker"
	t.Run(name, func(t *testing.T) {
		// the gated Filter with no broker should FlushAll events to /dev/null
		// and therefore produce no events in the file sink
		assert, require := assert.New(t), require.New(t)
		ctx := context.Background()

		b, gf, tmpDir := testBrokerWithGatedFilter(t, name, "test")
		gf.Broker = nil

		payload := &gated.Payload{
			ID: "event-1",
			Header: map[string]interface{}{
				"user": "alice",
				"tmz":  "EST",
			},
			Detail: map[string]interface{}{
				"file_name":   "file1.txt",
				"total_bytes": 1024,
			},
		}

		_, err := b.Send(ctx, eventlogger.EventType("test"), payload)
		require.NoError(err)

		err = gf.FlushAll(ctx)
		require.NoError(err)

		files, err := ioutil.ReadDir(tmpDir)
		require.NoError(err)
		assert.Len(files, 0)
	})
}

func TestGatedFilter_Now(t *testing.T) {
	t.Parallel()
	t.Run("default-now", func(t *testing.T) {
		assert := assert.New(t)
		gf := gated.Filter{}
		n := time.Now()
		got := gf.Now()
		assert.True(got.Equal(time.Now()) || got.Before(time.Now()))
		assert.True(got.Equal(n) || got.After(n))
	})
	t.Run("override-now", func(t *testing.T) {
		assert := assert.New(t)
		n := time.Now()
		gf := gated.Filter{
			NowFunc: func() time.Time { return n },
		}
		assert.Equal(n, gf.Now())
	})
}

func TestGatedFilter_Type(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	gf := gated.Filter{}
	assert.Equal(eventlogger.NodeTypeFilter, gf.Type())
}

func testBrokerWithGatedFilter(t *testing.T, testName string, eventType string) (*eventlogger.Broker, *gated.Filter, string) {
	t.Helper()
	require := require.New(t)
	require.NotEmpty(eventType)
	tmpDir, err := ioutil.TempDir("", testName)
	require.NoError(err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// Create a broker
	b := eventlogger.NewBroker()

	gf := &gated.Filter{
		Broker: b,
	}

	// Marshal to JSON
	n1 := &eventlogger.JSONFormatter{}
	// Send to FileSink
	n2 := &eventlogger.FileSink{Path: tmpDir, FileName: "file.log"}

	// Register the graph with the broker
	et := eventlogger.EventType(eventType)
	nodes := []eventlogger.Node{gf, n1, n2}
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
	return b, gf, tmpDir
}
