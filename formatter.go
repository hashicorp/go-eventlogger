package eventlogger

import (
	"bytes"
	"encoding/json"
	"time"
)

type JSONFormatter struct {
	nodes []Node
}

var _ LinkableNode = &JSONFormatter{}

func (w *JSONFormatter) Process(e *Event) (*Event, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(struct {
		CreatedAt time.Time
		EventType
		Payload interface{}
	}{
		e.CreatedAt,
		e.Type,
		e.Payload,
	})
	if err != nil {
		return nil, err
	}

	e.l.Lock()
	e.Formatted["json"] = buf.Bytes()
	e.l.Unlock()
	return e, nil
}

func (w *JSONFormatter) SetNext(nodes []Node) {
	w.nodes = nodes
}

func (w *JSONFormatter) Next() []Node {
	return w.nodes
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
