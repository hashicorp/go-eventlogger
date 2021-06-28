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
	NodeName    = "cloudevents-formatter" // NodeName defines the name of Formatter nodes
	SpecVersion = "1.0"                   // SpecVersion defines the cloudevents spec version supported
	TextIndent  = "  "                    // TextIndent defines the prefix/indent used when encoding FormatText
)

// ID defines an optional single function interface that Event Payloads may
// implement which returns the cloudevent ID for the event payload. If an Event
// Payload doesn't implement this optional interface then a unique ID is
// generated and used for the cloudevent's ID.
type ID interface {
	// ID returns the cloudevent ID
	ID() string
}

// Data defines an optional single function interface that Event Payloads may
// implement which returns the cloudevent data for the event payload. If an
// Event doesn't implement this optional interface then the entire Event Payload
// is used as the cloudevent data.
type Data interface {
	// Data returns the cloudevent Data.
	Data() interface{}
}

// CloudEvent defines type which is used when formatting cloudevents.
//
// For more info on the fields see: https://github.com/cloudevents/spec)
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

// Formatter is a Node which formats the Event as a CloudEvent in JSON
// format (See: https://github.com/cloudevents/spec)
type Formatter struct {
	// Source identifies the context where the cloudevents happen and is
	// required
	Source *url.URL

	// Schema is the JSON schema for the cloudevent data and is optional
	Schema *url.URL

	// Format defines the format created by the node.  If empty (unspecified),
	// FormatJSON will be used
	Format Format
}

var _ eventlogger.Node = &Formatter{}

func (f *Formatter) validate() error {
	const op = "cloudevents.(Formatter).validate"
	if f == nil {
		return fmt.Errorf("%s: missing formatter: %w", op, eventlogger.ErrInvalidParameter)
	}
	if f.Source == nil || f.Source.String() == "" {
		return fmt.Errorf("%s: missing source: %w", op, eventlogger.ErrInvalidParameter)
	}
	if err := f.Format.validate(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if f.Schema != nil && f.Schema.String() == "" {
		return fmt.Errorf("%s: an empty schema is not valid: %w", op, eventlogger.ErrInvalidParameter)
	}
	return nil
}

// Process formats the Event as a cloudevent and stores that formatted data in
// Event.Formatted with a key of "cloudevents-json" (cloudevents.FormatJSON)
func (f *Formatter) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	const op = "cloudevents.(Formatter).Process"
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("%s: invalid Formatter %w", op, err)
	}
	if e == nil {
		return nil, fmt.Errorf("%s: missing event: %w", op, eventlogger.ErrInvalidParameter)
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
		if id == "" {
			return nil, fmt.Errorf("%s: returned ID() is empty: %w", op, eventlogger.ErrInvalidParameter)
		}
	} else {
		var err error
		id, err = newId()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	var schema string
	if f.Schema != nil {
		schema = f.Schema.String()
	}

	ce := CloudEvent{
		ID:          id,
		Source:      f.Source.String(),
		SpecVersion: SpecVersion,
		Type:        string(e.Type),
		Data:        data,
		DataSchema:  schema,
		Time:        e.CreatedAt,
	}
	switch f.Format {
	case FormatJSON, FormatUnspecified:
		ce.DataContentType = DataContentTypeCloudEvents
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode(ce); err != nil {
			return nil, fmt.Errorf("%s: error formatting as JSON: %w", op, err)
		}
		e.FormattedAs(string(FormatJSON), buf.Bytes())
	case FormatText:
		ce.DataContentType = DataContentTypeText
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetIndent(TextIndent, TextIndent)
		if err := enc.Encode(ce); err != nil {
			return nil, fmt.Errorf("%s: error formatting as text: %w", op, err)
		}
		e.FormattedAs(string(FormatText), buf.Bytes())
	default:
		// this should be unreachable since f.validate() should catch this error
		// condition at the top of the function.
		return nil, fmt.Errorf("%s: %s is not a supported format: %w", op, f.Format, eventlogger.ErrInvalidParameter)
	}
	return e, nil
}

// Reopen is a no op
func (f *Formatter) Reopen() error {
	return nil
}

// Type describes the type of the node as a Formatter.
func (f *Formatter) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeFormatter
}

// Name returns a representation of the Formatter's name
func (f *Formatter) Name() string {
	return NodeName
}
