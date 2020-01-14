package eventlogger

import (
	"context"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	w := &JSONFormatter{}
	e := &Event{
		Formatted: make(map[string][]byte),
	}
	_, err := w.Process(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
}
