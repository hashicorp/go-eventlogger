package cloudevents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/eventlogger"
)

const (
	NodeName    = "cloudevents-formatter"
	SpecVersion = "1.0"
)

type ID interface {
	ID() string
}

type Data interface {
	Data() string
}

type CloudEvent struct {
	// ID identifies the event, cannot be an empty and is required.  The
	// combination of Source + ID must be unique.  Events with the same Source +
	// ID can be assumed to be duplicates by consumers
	ID string `json:"id"`

	// Source identifies the context in which the event happened, it is a
	// URI-reference, cannot be empty and is required.
	Source string `json:"source"`

	// SpecVersion defines the version of CloudEvents that the event is using,
	// it cannot be empty and is required.
	SpecVersion string `json:"specversion"`

	// Type defines the event's type, cannot be empty and is required.
	Type string `json:"type"`

	// Data may include domain-specific information about the event and is
	// optional.
	Data interface{} `json:"data,omitempty"`

	// DataContentType defines the content type of the event's data value and is
	// optional.  If present it must adhere to:
	// https://datatracker.ietf.org/doc/html/rfc2046
	DataContentType string `json:"datacontentype,omitempty"`

	// DataSchema is a URI-reference and is optional.
	DataSchema string `json:"dataschema,omitempty"`

	// Time is in format RFC 3339 (the default for time.Time) and is optional
	Time time.Time `json:"time,omitempty"`
}

// JSONFormatter is a Node which formats the Event as a CloudEvent in JSON
// format (See: https://github.com/cloudevents/spec)
type JSONFormatter struct {
	// Source identifies the context where the events happen and is required.
	Source url.URL

	// Schema is the JSON schema for the event data (aka payload) and is optional
	Schema *url.URL

	// Format defines the format created by the node.  If empty (unspecified),
	// FormatJSON will be used.
	Format Format
}

var _ eventlogger.Node = &JSONFormatter{}

func (f *JSONFormatter) validate() error {
	const op = "cloudevents.(JSONFormatter).validate"
	if f.Source.String() == "" {
		return fmt.Errorf("%s: missing source: %w", op, eventlogger.ErrInvalidParameter)
	}
	if err := f.Format.validate(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Process formats the Event as a cloudevent and stores that formatted data in
// Event.Formatted with a key of "cloudevents-json" (cloudevents.FormatJSON)
func (f *JSONFormatter) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	const op = "cloudevents.(JSONFormatter).Process"
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("%s: invalid JSONFormatter %w", op, err)
	}

	var data interface{}
	if i, ok := e.Payload.(Data); ok {
		data = i.Data()
	} else {
		data = e.Payload
	}
	var id string
	if i, ok := e.Payload.(ID); ok {
		id = i.ID()
	} else {
		var err error
		id, err = newId()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	ce := CloudEvent{
		ID:          id,
		Source:      f.Source.String(),
		SpecVersion: SpecVersion,
		Type:        string(e.Type),
		Data:        data,
		DataSchema:  f.Schema.String(),
		Time:        e.CreatedAt,
	}
	switch f.Format {
	case FormatText, FormatUnspecified:
		ce.DataContentType = DataContentTypeText

	default:
		ce.DataContentType = DataContentTypeCloudEvents

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode(ce); err != nil {
			return nil, err
		}
		e.FormattedAs(string(FormatJSON), buf.Bytes())
	}
	return e, nil
}

// Reopen is a no op
func (f *JSONFormatter) Reopen() error {
	return nil
}

// Type describes the type of the node as a Formatter.
func (f *JSONFormatter) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeFormatter
}

// Name returns a representation of the Formatter's name
func (f *JSONFormatter) Name() string {
	return NodeName
}
