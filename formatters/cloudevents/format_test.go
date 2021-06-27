package cloudevents

import (
	"testing"

	"github.com/hashicorp/eventlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormat_validate(t *testing.T) {
	tests := []struct {
		name      string
		f         Format
		wantErrIs error
	}{
		{
			name:      "invalid-format",
			f:         "invalid",
			wantErrIs: eventlogger.ErrInvalidParameter,
		},
		{
			name: "valid-empty-format",
		},
		{
			name: "valid-text-format",
			f:    FormatText,
		},
		{
			name: "valid-json-format",
			f:    FormatJSON,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			err := tt.f.validate()
			if tt.wantErrIs != nil {
				require.Error(err)
				assert.ErrorIs(err, tt.wantErrIs)
				return
			}
			require.NoError(err)
		})
	}
}

func TestFormat_convertToDataContentType(t *testing.T) {
	tests := []struct {
		name string
		f    Format
		want string
	}{
		{
			name: "format-json",
			f:    FormatJSON,
			want: DataContentTypeCloudEvents,
		},
		{
			name: "empty-format",
			want: DataContentTypeCloudEvents,
		},
		{
			name: "text-format",
			f:    FormatText,
			want: DataContentTypeText,
		},
		{
			name: "whatever-format",
			f:    "whatever",
			want: DataContentTypeText,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got := tt.f.convertToDataContentType()
			assert.Equal(tt.want, got)
		})
	}
}
