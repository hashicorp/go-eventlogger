package eventlogger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileSink writes the []byte representation of an Event to a file
// as a string.
type FileSink struct {
	// Path is the complete path of the log file
	Path string

	// FileName is the name of the log file
	FileName string

	// Mode is the file's mode and permission bits
	Mode os.FileMode

	// LastCreated represents the creation time of the latest log
	LastCreated time.Time

	// MaxBytes is the maximum number of desired bytes for a log file
	MaxBytes int

	// BytesWritten is the number of bytes written in the current log file
	BytesWritten int64

	// MaxFiles is the maximum number of old files to keep before removing them
	MaxFiles int

	// MaxDuration is the maximum duration allowed between each file rotation
	MaxDuration time.Duration

	// Format specifies the format the []byte representation is formatted in
	// Defaults to "json"
	Format string

	f *os.File
	l sync.Mutex
}

var _ Node = &FileSink{}

const defaultMode = 0600

func (fs *FileSink) Type() NodeType {
	return NodeTypeSink
}

func (fs *FileSink) Process(ctx context.Context, e *Event) (*Event, error) {
	format := fs.Format
	if format == "" {
		format = "json"
	}
	e.l.RLock()
	val, ok := e.Formatted[format]
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

	// Check for last contact, rotate if necessary and able
	if err := fs.rotate(); err != nil {
		return nil, err
	}

	if n, err := reader.WriteTo(fs.f); err == nil {
		// Sinks are leafs, so do not return the event, since nothing more can
		// happen to it downstream.
		fs.BytesWritten += int64(n)
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

	if fs.f != nil {
		// Ensure file still exists
		_, err := os.Stat(fs.f.Name())
		if os.IsNotExist(err) {
			fs.f = nil
		}
	}

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

func (fs *FileSink) open() error {
	mode := fs.Mode
	if mode == 0 {
		mode = defaultMode
	}

	if err := os.MkdirAll(filepath.Dir(fs.Path), mode); err != nil {
		return err
	}

	createTime := time.Now()
	// New file name as the format:
	// file rotation enabled: filename-timestamp.extension
	// file rotation disabled: filename.extension
	newfileName := fs.newFileName(createTime)
	newfilePath := filepath.Join(fs.Path, newfileName)

	var err error
	fs.f, err = os.OpenFile(newfilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, mode)
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

	fs.LastCreated = createTime
	fs.BytesWritten = 0
	return nil
}

func (fs *FileSink) rotate() error {
	// Get the time from the last point of contact
	elapsed := time.Since(fs.LastCreated)
	if (fs.BytesWritten >= int64(fs.MaxBytes) && (fs.MaxBytes > 0)) ||
		((elapsed > fs.MaxDuration) && (fs.MaxDuration > 0)) {

		fs.f.Close()
		if err := fs.pruneFiles(); err != nil {
			return err
		}
		return fs.open()
	}

	return nil
}

func (fs *FileSink) pruneFiles() error {
	if fs.MaxFiles == 0 {
		return nil
	}

	// get all the files that match the log file pattern
	pattern := fs.fileNamePattern()
	globExpression := filepath.Join(fs.Path, fmt.Sprintf(pattern, "*"))
	matches, err := filepath.Glob(globExpression)
	if err != nil {
		return err
	}

	// Stort the strings as filepath.Glob does not publicly guarantee that files
	// are sorted, so here we add an extra defensive sort.
	sort.Strings(matches)

	stale := len(matches) - fs.MaxFiles
	for i := 0; i < stale; i++ {
		if err := os.Remove(matches[i]); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FileSink) fileNamePattern() string {
	// Extract file extension
	ext := filepath.Ext(fs.FileName)
	if ext == "" {
		ext = ".log"
	}

	// Add format string between file and extension
	return strings.TrimSuffix(fs.FileName, ext) + "-%s" + ext
}

func (fs *FileSink) newFileName(createTime time.Time) string {
	var newFileName string
	if fs.rotateEnabled() {
		pattern := fs.fileNamePattern()
		newFileName = fmt.Sprintf(pattern, strconv.FormatInt(createTime.UnixNano(), 10))
	} else {
		newFileName = fs.FileName
	}
	return newFileName
}

func (fs *FileSink) rotateEnabled() bool {
	return fs.MaxBytes > 0 || fs.MaxDuration != 0
}
