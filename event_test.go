package eventlogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent_FormattedAs(t *testing.T) {
	tests := []struct {
		name            string
		e               *Event
		formatType      string
		formatttedValue []byte
		want            []byte
	}{
		{
			name:            "text",
			e:               &Event{Formatted: map[string][]byte{}},
			formatType:      "text",
			formatttedValue: []byte("text-format"),
			want:            []byte("text-format"),
		},
		{
			name:            "nil-formatted-map",
			e:               &Event{Formatted: nil},
			formatType:      "text",
			formatttedValue: []byte("text-format"),
			want:            []byte("text-format"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			tt.e.FormattedAs(tt.formatType, tt.formatttedValue)
			got, ok := tt.e.Format(tt.formatType)
			require.Truef(ok, "format %s not found", tt.formatType)
			assert.Equal(tt.want, got)
		})
	}
}

func TestEvent_Format(t *testing.T) {
	tests := []struct {
		name       string
		e          *Event
		formatType string
		want       []byte
		wantOk     bool
	}{
		{
			name:       "found",
			e:          &Event{Formatted: map[string][]byte{"found": []byte("found")}},
			formatType: "found",
			want:       []byte("found"),
			wantOk:     true,
		},
		{
			name:       "not-ok-with-nil-Formatted-map",
			e:          &Event{},
			formatType: "not-ok",
			want:       nil,
			wantOk:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got, gotOk := tt.e.Format(tt.formatType)
			assert.Equal(tt.wantOk, gotOk)
			assert.Equal(tt.want, got)
		})
	}
}
