package eventlogger

import (
	"errors"
	"os"
)

//----------------------------------------------------------
// Node

type Node interface {
	Process(e *Envelope) error
}

type LinkableNode interface {
	Node
	SetNext([]Node)
	Next() []Node
}

//----------------------------------------------------------
// Filter

// Predicate returns true if we want to keep the Envelope.
type Predicate func(e *Envelope) (bool, error)

// Filter
type Filter struct {
	nodes []Node

	Predicate Predicate
}

var DiscardEvent = errors.New("DiscardEvent")

func (f *Filter) Process(e *Envelope) error {

	// Use the predicate to see if we want to keep the event.
	keep, err := f.Predicate(e)
	if err != nil {
		return err
	}
	if !keep {
		return DiscardEvent
	}
	return nil
}

func (f *Filter) SetNext(nodes []Node) {
	f.nodes = nodes
}

func (f *Filter) Next() []Node {
	return f.nodes
}

//----------------------------------------------------------
// ByteWriter

// ByteMarshaller turns an Envelope into a slice of bytes suitable for being
// persisted.
type ByteMarshaller func(e *Envelope) ([]byte, error)

// ByteWriter
type ByteWriter struct {
	nodes []Node

	Marshaller ByteMarshaller
}

func (w *ByteWriter) Process(e *Envelope) error {

	bytes, err := w.Marshaller(e)
	if err != nil {
		return err
	}

	e.Marshalled = bytes
	return nil
}

func (w *ByteWriter) SetNext(nodes []Node) {
	w.nodes = nodes
}

func (w *ByteWriter) Next() []Node {
	return w.nodes
}

//----------------------------------------------------------
// FileSink

// FileSink writes the []byte representation of an Envelope to a file
// as a string.
type FileSink struct {
	FilePath string
}

func (fs *FileSink) Process(e *Envelope) error {

	bytes, ok := (e.Marshalled).([]byte)
	if !ok {
		return errors.New("Envelope is not writable to a FileSink")
	}

	f, err := os.OpenFile(fs.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(string(bytes)); err != nil {
		return err
	}
	if _, err = f.WriteString("\n"); err != nil {
		return err
	}

	return nil
}
