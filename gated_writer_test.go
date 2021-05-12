package eventlogger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/require"
)

func TestGatedFilter_simple(t *testing.T) {
	require := require.New(t)
	gw := eventlogger.GatedFilter{}

	events := []*eventlogger.Event{
		{
			Type:      "test",
			CreatedAt: time.Now(),
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
			CreatedAt: time.Now(),
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
	for _, e := range events {
		got, err := gw.Process(context.Background(), e)
		require.NoError(err)
		require.Empty(got)
	}

	got, err := gw.Process(context.Background(), &eventlogger.Event{
		Type:      "test",
		CreatedAt: time.Now(),
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
	j, err := json.Marshal(got)
	require.NoError(err)
	fmt.Println(string(j))

}
