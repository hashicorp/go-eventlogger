package eventlogger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

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

func (fs *FileSink) Process(ctx context.Context, e *Event) (*Event, error) {
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
	_ = fs.f.Close()
	fs.f = nil

	if err := fs.open(); err != nil {
		return nil, err
	}

	_, _ = reader.Seek(0, io.SeekStart)
	_, err := reader.WriteTo(fs.f)
	return nil, err
}

func (fs *FileSink) Reopen() error {
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

func (fs *FileSink) Name() string {
	return fmt.Sprintf("sink:%s", fs.Path)
}
