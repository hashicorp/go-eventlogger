package eventlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"time"
)

const (
	JSONFormat = "json"
)

type JSONFormatter struct{}

var _ Node = &JSONFormatter{}

func (w *JSONFormatter) Process(ctx context.Context, e *Event) (*Event, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(struct {
		CreatedAt time.Time `json:"created_at"`
		EventType `json:"event_type"`
		Payload   interface{} `json:"payload"`
	}{
		e.CreatedAt,
		e.Type,
		e.Payload,
	})
	if err != nil {
		return nil, err
	}

	e.l.Lock()
	e.Formatted[JSONFormat] = buf.Bytes()
	e.l.Unlock()
	return e, nil
}

func (w *JSONFormatter) Reopen() error {
	return nil
}

func (w *JSONFormatter) Type() NodeType {
	return NodeTypeFormatter
}

func (w *JSONFormatter) Name() string {
	return "JSONFormatter"
}
