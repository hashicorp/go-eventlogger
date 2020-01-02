package eventlogger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//----------------------------------------------------------
// Node

type NodeType int

const (
	_ NodeType = iota
	NodeTypeFilter
	NodeTypeFormatter
	NodeTypeSink
)

// A Node in a Graph
type Node interface {
	// Process does something with the Event: filter, redaction,
	// marshalling, persisting.
	Process(e *Event) (*Event, error)
	// Reload is used to re-read any config stored externally
	// and to close and reopen files, e.g. for log rotation.
	Reload() error
	Type() NodeType
}

// A LinkableNode is a Node that has downstream children.  Nodes
// that are *not* LinkableNodes are Leafs.
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

// Predicate returns true if we want to keep the Event.
type Predicate func(e *Event) (bool, error)

// Filter
type Filter struct {
	nodes []Node

	Predicate Predicate
}

var _ LinkableNode = &Filter{}

func (f *Filter) Process(e *Event) (*Event, error) {

	// Use the predicate to see if we want to keep the event.
	keep, err := f.Predicate(e)
	if err != nil {
		return nil, err
	}
	if !keep {
		// Return nil to signal that the event should be discarded.
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

func (f *Filter) Reload() error {
	return nil
}

func (f *Filter) Type() NodeType {
	return NodeTypeFilter
}

//----------------------------------------------------------
// ByteWriter

// ByteMarshaller turns an Event into a slice of bytes suitable for being
// persisted.
type ByteMarshaller func(e *Event) ([]byte, error)

// JSONMarshaller marshals the envelope into JSON.  For now, it just
// does the Payload field.
var JSONMarshaller = func(e *Event) ([]byte, error) {
	w := &bytes.Buffer{}
	enc := json.NewEncoder(w)
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

	return w.Bytes(), nil
}

// ByteWriter
type ByteWriter struct {
	nodes []Node

	Marshaller ByteMarshaller
}

var _ LinkableNode = &ByteWriter{}

func (w *ByteWriter) Process(e *Event) (*Event, error) {
	bytes, err := w.Marshaller(e)
	if err != nil {
		return nil, err
	}

	e.l.Lock()
	e.Formatted["json"] = bytes
	e.l.Unlock()
	return e, nil
}

func (w *ByteWriter) SetNext(nodes []Node) {
	w.nodes = nodes
}

func (w *ByteWriter) Next() []Node {
	return w.nodes
}

func (w *ByteWriter) Reload() error {
	return nil
}

func (w *ByteWriter) Type() NodeType {
	return NodeTypeFormatter
}

//----------------------------------------------------------
// FileSink

// FileSink writes the []byte representation of an Event to a file
// as a string.
type FileSink struct {
	Path string
	Mode os.FileMode
	f    *os.File
	l    sync.Mutex
}

var _ Node = &FileSink{}

func (fs *FileSink) Type() NodeType {
	return NodeTypeSink
}

const defaultMode = 0600

func (fs *FileSink) open() error {
	mode := fs.Mode
	if mode == 0 {
		mode = defaultMode
	}

	if err := os.MkdirAll(filepath.Dir(fs.Path), mode); err != nil {
		return err
	}

	var err error
	fs.f, err = os.OpenFile(fs.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}

	// Change the file mode in case the log file already existed. We special
	// case /dev/null since we can't chmod it and bypass if the mode is zero
	switch fs.Path {
	case "/dev/null":
	default:
		if fs.Mode != 0 {
			err = os.Chmod(fs.Path, fs.Mode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FileSink) Process(e *Event) (*Event, error) {
	e.l.RLock()
	val, ok := e.Formatted["json"]
	e.l.RUnlock()
	if !ok {
		return nil, errors.New("event was not marshaled")
	}
	reader := bytes.NewReader(val)

	fs.l.Lock()
	defer fs.l.Unlock()

	if fs.f == nil {
		err := fs.open()
		if err != nil {
			return nil, err
		}
	}

	if _, err := reader.WriteTo(fs.f); err == nil {
		// Sinks are leafs, so do not return the event, since nothing more can
		// happen to it downstream.
		return nil, nil
	} else if fs.Path == "stdout" {
		return nil, err
	}

	// If writing to stdout there's no real reason to think anything would have
	// changed so return above. Otherwise, opportunistically try to re-open the
	// FD, once per call.
	fs.f.Close()
	fs.f = nil

	if err := fs.open(); err != nil {
		return nil, err
	}

	_, _ = reader.Seek(0, io.SeekStart)
	_, err := reader.WriteTo(fs.f)
	return nil, err
}

func (fs *FileSink) Reload() error {
	switch fs.Path {
	case "stdout", "discard":
		return nil
	}

	fs.l.Lock()
	defer fs.l.Unlock()

	if fs.f == nil {
		return fs.open()
	}

	err := fs.f.Close()
	// Set to nil here so that even if we error out, on the next access open()
	// will be tried
	fs.f = nil
	if err != nil {
		return err
	}

	return fs.open()
}
