package eventlogger

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriterSink_Process(t *testing.T) {
	ctx := context.Background()

	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}

	tests := []struct {
		name    string
		writer  io.ReadWriter
		e       *Event
		want    string
		wantErr bool
	}{
		{
			name:   "simple",
			writer: &bytes.Buffer{},
			e:      event,
			want:   "first",
		},
		{
			name:    "nil-writer",
			writer:  nil,
			e:       event,
			wantErr: true,
		},
		{
			name:    "nil-event",
			writer:  &bytes.Buffer{},
			e:       nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := WriterSink{
				Writer: tt.writer,
			}
			_, err := s.Process(ctx, tt.e)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got, err := ioutil.ReadAll(tt.writer)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("expected %s and got: %s", tt.want, string(got))
			}
		})
	}
	t.Run("stdout", func(t *testing.T) {
		s := WriterSink{
			Writer: os.Stdout,
		}
		_, err := s.Process(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}
	})
	t.Run("stderr", func(t *testing.T) {
		s := WriterSink{
			Writer: os.Stderr,
		}
		_, err := s.Process(ctx, event)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}
	})
}
