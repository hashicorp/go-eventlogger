package cloudevents

import (
	"context"
	"net/url"
	"testing"

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

	tests := []struct {
		name            string
		f               *Formatter
		e               *eventlogger.Event
		format          Format
		wantFormatted   string
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
			gotFormatted, ok := gotEvent.Format(tt.wantFormatted)
			require.True(ok)
			assert.Equal(tt.wantFormatted, gotFormatted)
		})
	}
}
