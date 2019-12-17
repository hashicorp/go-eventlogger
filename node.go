package eventlogger

import (
	"errors"
	"os"
)

//----------------------------------------------------------
// Node

type Node interface {
	Process(e *Envelope) (*Envelope, error)
}

type LinkableNode interface {
	Node
	SetNext([]Node)
	Next() []Node
}

// LinkNodes is a convenience function that connects
// Nodes together into a linked list. All of the nodes except the
// last one must be LinkableNodes
func LinkNodes(nodes []Node) ([]Node, error) {

	num := len(nodes)
	if num < 2 {
		return nodes, nil
	}

	for i := 0; i < num-1; i++ {
		ln, ok := nodes[i].(LinkableNode)
		if !ok {
			return nil, errors.New("Node is not Linkable")
		}
		ln.SetNext([]Node{nodes[i+1]})
	}

	return nodes, nil
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

func (f *Filter) Process(e *Envelope) (*Envelope, error) {

	// Use the predicate to see if we want to keep the event.
	keep, err := f.Predicate(e)
	if err != nil {
		return nil, err
	}
	if !keep {
		return nil, nil
	}

	// return the event
	return e, nil
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

func (w *ByteWriter) Process(e *Envelope) (*Envelope, error) {

	bytes, err := w.Marshaller(e)
	if err != nil {
		return nil, err
	}

	e.Marshalled = bytes
	return e, nil
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

func (fs *FileSink) Process(e *Envelope) (*Envelope, error) {

	bytes, ok := (e.Marshalled).([]byte)
	if !ok {
		return nil, errors.New("Envelope is not writable to a FileSink")
	}

	f, err := os.OpenFile(fs.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err = f.WriteString(string(bytes)); err != nil {
		return nil, err
	}
	if _, err = f.WriteString("\n"); err != nil {
		return nil, err
	}

	// Do not return the event, since nothing more can happen to it downstream.
	return nil, nil
}
