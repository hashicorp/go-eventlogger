package cloudevents

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatterFilter_Type(t *testing.T) {
	t.Parallel()
	t.Run("assert-type", func(t *testing.T) {
		assert := assert.New(t)
		f := FormatterFilter{}
		assert.Equal(eventlogger.NodeTypeFormatterFilter, f.Type())
	})
}

func TestFormatterFilter_Name(t *testing.T) {
	t.Parallel()
	t.Run("assert-name", func(t *testing.T) {
		assert := assert.New(t)
		f := FormatterFilter{}
		assert.Equal(NodeName, f.Name())
	})
}

func TestFormatterFilter_Reopen(t *testing.T) {
	t.Parallel()
	t.Run("assert-no-error", func(t *testing.T) {
		assert := assert.New(t)
		f := FormatterFilter{}
		assert.NoError(f.Reopen())
	})
}

func TestFormatterFilter_Process(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testURL, err := url.Parse("https://localhost")
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name            string
		f               *FormatterFilter
		e               *eventlogger.Event
		format          Format
		wantCloudEvent  *Event
		wantText        string
		wantIsError     error
		wantErrContains string
	}{
		{
			name:            "missing-formatter-filter",
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing formatter filter",
		},
		{
			name:            "missing-source-nil",
			f:               &FormatterFilter{},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing source",
		},
		{
			name: "missing-source-empty",
			f: &FormatterFilter{
				Source: &url.URL{},
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing source",
		},
		{
			name: "invalid-format",
			f: &FormatterFilter{
				Source: testURL,
				Format: "invaid",
				Schema: testURL,
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "not a valid format",
		},
		{
			name: "empty-schema",
			f: &FormatterFilter{
				Source: testURL,
				Format: FormatJSON,
				Schema: &url.URL{},
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "empty schema",
		},
		{
			name: "missing-event",
			f: &FormatterFilter{
				Source: testURL,
			},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing event",
		},
		{
			name: "simple-JSON",
			f: &FormatterFilter{
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
			wantCloudEvent: &Event{
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
			name: "filter-no-error",
			f: &FormatterFilter{
				Source: testURL,
				Schema: testURL,
				Format: FormatJSON,
				Predicate: func(ctx context.Context, cloudevent interface{}) (bool, error) {
					return false, nil
				},
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format: FormatJSON,
		},
		{
			name: "filter-with-error",
			f: &FormatterFilter{
				Source: testURL,
				Schema: testURL,
				Format: FormatJSON,
				Predicate: func(ctx context.Context, cloudevent interface{}) (bool, error) {
					return false, eventlogger.ErrInvalidParameter
				},
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format:      FormatJSON,
			wantIsError: eventlogger.ErrInvalidParameter,
		},
		{
			name: "optional-interfaces",
			f: &FormatterFilter{
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
			wantCloudEvent: &Event{
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
			name: "optional-interfaces",
			f: &FormatterFilter{
				Source: testURL,
				Format: FormatJSON,
			},
			e: &eventlogger.Event{
				Type:      "optional-interfaces",
				CreatedAt: now,
				Payload: &testOptionalInterfaces{
					payload: map[string]interface{}{
						"id": "",
						"data": map[string]interface{}{
							"name": "alice",
							"dob":  now,
						},
					},
				},
			},
			format:          FormatJSON,
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "returned ID() is empty",
		},
		{
			name: "simple-Text",
			f: &FormatterFilter{
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
			wantCloudEvent: &Event{
				Source:          testURL.String(),
				DataSchema:      testURL.String(),
				SpecVersion:     SpecVersion,
				Type:            "test",
				Data:            "test-string",
				DataContentType: "text/plain",
				Time:            now,
			},
			wantText: `{
  "id": "%s",
  "source": "https://localhost",
  "specversion": "1.0",
  "type": "test",
  "data": "test-string",
  "datacontentype": "text/plain",
  "dataschema": "https://localhost",
  "time": %s
}
`,
		},
		{
			name: "simple-signer",
			f: &FormatterFilter{
				Source: testURL,
				Schema: testURL,
				Format: FormatJSON,
				Signer: func(c context.Context, b []byte) (string, error) {
					return "signature", nil
				},
				SignEventTypes: []string{"test"},
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format: FormatJSON,
			wantCloudEvent: &Event{
				Source:          testURL.String(),
				DataSchema:      testURL.String(),
				SpecVersion:     SpecVersion,
				Type:            "test",
				Data:            "test-string",
				DataContentType: "application/cloudevents",
				Time:            now,
				SerializedHmac:  "signature",
			},
		},
		{
			name: "simple-signer-without-signature",
			f: &FormatterFilter{
				Source: testURL,
				Schema: testURL,
				Format: FormatJSON,
				Signer: func(c context.Context, b []byte) (string, error) {
					return "signature", nil
				},
				SignEventTypes: []string{"not-match-type"},
			},
			e: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   "test-string",
			},
			format: FormatJSON,
			wantCloudEvent: &Event{
				Source:          testURL.String(),
				DataSchema:      testURL.String(),
				SpecVersion:     SpecVersion,
				Type:            "test",
				Data:            "test-string",
				DataContentType: "application/cloudevents",
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
			if tt.wantCloudEvent == nil {
				assert.Nil(gotEvent)
				return
			}
			gotFormatted, ok := gotEvent.Format(string(tt.format))
			require.True(ok)
			var gotCloudEvent Event
			require.NoError(json.Unmarshal(gotFormatted, &gotCloudEvent))
			if tt.wantCloudEvent.ID == "" {
				tt.wantCloudEvent.ID = gotCloudEvent.ID
			}
			if gotCloudEvent.SerializedHmac != "" {
				hmac := tt.wantCloudEvent.SerializedHmac
				tt.wantCloudEvent.SerializedHmac = ""
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.Encode(tt.wantCloudEvent)
				require.NoError(err)
				tt.wantCloudEvent.Serialized = base64.RawURLEncoding.EncodeToString(buf.Bytes())
				tt.wantCloudEvent.SerializedHmac = hmac
			}
			var wantJSON []byte
			switch tt.format {
			case FormatJSON:
				wantJSON, err = json.Marshal(tt.wantCloudEvent)
			case FormatText:
				// test the raw JSON
				jsonTime, err := gotCloudEvent.Time.MarshalJSON()
				require.NoError(err)
				wantRawText := []byte(fmt.Sprintf(tt.wantText, gotCloudEvent.ID, jsonTime))
				assert.Equal(string(wantRawText), string(gotFormatted))

				// test the marshaled JSON
				wantJSON, err = json.MarshalIndent(tt.wantCloudEvent, TextIndent, TextIndent)
				require.NoError(err)
			}
			require.NoError(err)
			assert.JSONEq(string(wantJSON), string(gotFormatted))
		})
	}
}

func TestFormatterFilter_Rotate(t *testing.T) {
	tests := []struct {
		name            string
		f               *FormatterFilter
		s               Signer
		wantIsError     error
		wantErrContains string
	}{
		{
			name:            "missing-signer",
			f:               &FormatterFilter{},
			wantIsError:     eventlogger.ErrInvalidParameter,
			wantErrContains: "missing signer",
		},
		{
			name: "valid",
			f:    &FormatterFilter{},
			s: func(c context.Context, b []byte) (string, error) {
				return "signature", nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			require.Nil(tt.f.Signer)
			err := tt.f.Rotate(tt.s)
			if tt.wantIsError != nil {
				require.Error(err)
				assert.ErrorIs(err, tt.wantIsError)
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			assert.NotNil(tt.f.Signer)
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
