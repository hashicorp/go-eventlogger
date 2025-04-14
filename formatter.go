// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	JSONFormat  = "json"
	ProtoFormat = "proto"
)

// JSONFormatter is a Formatter Node which formats the Event as JSON.
type JSONFormatter struct{}

var _ Node = &JSONFormatter{}

// Process formats the Event as JSON and stores that formatted data in
// Event.Formatted with a key of "json"
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

	e.FormattedAs(JSONFormat, buf.Bytes())
	return e, nil
}

// Reopen is a no op
func (w *JSONFormatter) Reopen() error {
	return nil
}

// Type describes the type of the node as a Formatter.
func (w *JSONFormatter) Type() NodeType {
	return NodeTypeFormatter
}

// Name returns a representation of the Formatter's name
func (w *JSONFormatter) Name() string {
	return "JSONFormatter"
}

type ProtoFormatter struct{}

var _ Node = &ProtoFormatter{}

// Process serializes the event's Payload using proto.Marshal.
// The Payload must implement proto.Message; otherwise, an error is returned.
func (f *ProtoFormatter) Process(_ context.Context, e *Event) (*Event, error) {
	msg, ok := e.Payload.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("payload is not proto.Message: %T", e.Payload)
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	e.FormattedAs(ProtoFormat, data)
	return e, nil
}

// Reopen is a no op
func (f *ProtoFormatter) Reopen() error {
	return nil
}

// Type describes the type of the node as a Formatter.
func (f *ProtoFormatter) Type() NodeType {
	return NodeTypeFormatter
}

// Name returns a representation of the Formatter's name
func (f *ProtoFormatter) Name() string {
	return "ProtoFormatter"
}
