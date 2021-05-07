package eventlogger

import (
	"bytes"
	"context"
	"errors"
	"io"
)

// WriterSink writes the []byte respresentation of an Event to an io.Writer as a
// string.  WriterSink allows you to define sinks for any io.Writer which
// includes os.Stdout and os.Stderr
type WriterSink struct {
	// Format specifies the format the []byte representation is formatted in
	// Defaults to "json"
	Format string

	// Writer is the io.Writer used when writing Events
	Writer io.Writer
}

// Reopen does nothing for WriterSinks.  The can not be rotated.
func (fs *WriterSink) Reopen() error { return nil }

// Type defines a WriterSink as a NodeTypeSink
func (fs *WriterSink) Type() NodeType {
	return NodeTypeSink
}

// Process will Write the event to the WriterSink
func (fs *WriterSink) Process(ctx context.Context, e *Event) (*Event, error) {
	if fs.Writer == nil {
		return nil, errors.New("sink writer is nil")
	}
	if e == nil {
		return nil, errors.New("event is nil")
	}

	format := fs.Format
	if fs.Format == "" {
		format = "json"
	}
	e.l.RLock()
	val, ok := e.Formatted[format]
	e.l.RUnlock()
	if !ok {
		return nil, errors.New("event was not marshaled")
	}
	reader := bytes.NewReader(val)

	if _, err := reader.WriteTo(fs.Writer); err != nil {
		return nil, err
	}

	// Sinks are leafs, so do not return the event, since nothing more can
	// happen to it downstream.
	return nil, nil
}
