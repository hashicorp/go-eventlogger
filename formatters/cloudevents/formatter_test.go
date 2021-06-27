package cloudevents

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_Type(t *testing.T) {
	t.Parallel()
	t.Run("assert-type", func(t *testing.T) {
		assert := assert.New(t)
		f := Formatter{}
		assert.Equal(eventlogger.NodeTypeFormatter, f.Type())
	})
}

func TestFormatter_Name(t *testing.T) {
	t.Parallel()
	t.Run("assert-name", func(t *testing.T) {
		assert := assert.New(t)
		f := Formatter{}
		assert.Equal(NodeName, f.Name())
	})
}

func TestFormatter_Reopen(t *testing.T) {
	t.Parallel()
	t.Run("assert-no-error", func(t *testing.T) {
		assert := assert.New(t)
		f := Formatter{}
		assert.NoError(f.Reopen())
	})
}

func TestFormatter_Process(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testURL, err := url.Parse("https://localhost")
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name            string
		f               *Formatter
		e               *eventlogger.Event
		format          Format
		wantCloudEvent  CloudEvent
		wantIsError     error
		wantErrContains string
	}{
		{
			name:            "missing-formatter",
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing formatter",
		},
		{
			name:            "missing-source-nil",
			f:               &Formatter{},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing source",
		},
		{
			name: "missing-source-empty",
			f: &Formatter{
				Source: &url.URL{},
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing source",
		},
		{
			name: "invalid-format",
			f: &Formatter{
				Source: testURL,
				Format: "invaid",
				Schema: testURL,
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "not a valid format",
		},
		{
			name: "empty-schema",
			f: &Formatter{
				Source: testURL,
				Format: FormatJSON,
				Schema: &url.URL{},
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "empty schema",
		},
		{
			name: "missing-event",
			f: &Formatter{
				Source: testURL,
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing event",
		},
		{
			name: "simple-JSON",
			f: &Formatter{
				Source: testURL,
				Schema: testURL,
				Format: FormatJSON,
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format: FormatJSON,
			wantCloudEvent: CloudEvent{
				Source:          testURL.String(),
				DataSchema:      testURL.String(),
				SpecVersion:     SpecVersion,
				Type:            "test",
				Data:            "test-string",
				DataContentType: "application/cloudevents",
				Time:            now,
			},
		},
		{
			name: "optional-interfaces",
			f: &Formatter{
				Source: testURL,
				Format: FormatJSON,
			},
			e: &eventlogger.Event{
				Type:      "optional-interfaces",
				CreatedAt: now,
				Payload: &testOptionalInterfaces{
					payload: map[string]interface{}{
						"id": "test-id",
						"data": map[string]interface{}{
							"name": "alice",
							"dob":  now,
						},
					},
				},
			},
			format: FormatJSON,
			wantCloudEvent: CloudEvent{
				ID:          "test-id",
				Source:      testURL.String(),
				SpecVersion: SpecVersion,
				Type:        "optional-interfaces",
				Data: map[string]interface{}{
					"name": "alice",
					"dob":  now,
				},
				DataContentType: "application/cloudevents",
				Time:            now,
			},
		},
		{
			name: "simple-Text",
			f: &Formatter{
				Source: testURL,
				Schema: testURL,
				Format: FormatText,
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format: FormatText,
			wantCloudEvent: CloudEvent{
				Source:          testURL.String(),
				DataSchema:      testURL.String(),
				SpecVersion:     SpecVersion,
				Type:            "test",
				Data:            "test-string",
				DataContentType: "text/plain",
				Time:            now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)

			gotEvent, err := tt.f.Process(ctx, tt.e)
			if tt.wantIsError != nil {
				require.Error(err)
				assert.Nil(gotEvent)
				assert.ErrorIs(err, tt.wantIsError)
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			gotFormatted, ok := gotEvent.Format(string(tt.format))
			require.True(ok)
			var gotCloudEvent CloudEvent
			require.NoError(json.Unmarshal(gotFormatted, &gotCloudEvent))
			if tt.wantCloudEvent.ID == "" {
				tt.wantCloudEvent.ID = gotCloudEvent.ID
			}
			var wantJSON []byte
			switch tt.format {
			case FormatJSON:
				wantJSON, err = json.Marshal(tt.wantCloudEvent)
			case FormatText:
				wantJSON, err = json.MarshalIndent(tt.wantCloudEvent, IndentWith, IndentWith)
			}
			require.NoError(err)
			assert.JSONEq(string(wantJSON), string(gotFormatted))
			t.Log(string(gotFormatted))
		})
	}
}

type testOptionalInterfaces struct {
	payload map[string]interface{}
}

func (t *testOptionalInterfaces) ID() string {
	return t.payload["id"].(string)
}

func (t *testOptionalInterfaces) Data() interface{} {
	return t.payload["data"].(interface{})
}
