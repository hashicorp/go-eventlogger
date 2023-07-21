// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestFileSink_NewDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	sinkDir := filepath.Join(tmpDir, "file_sink")

	fs := FileSink{
		Path:     sinkDir,
		FileName: "audit.log",
	}

	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := []string{"audit.log"}
	files, _ := os.ReadDir(sinkDir)
	got := []string{}
	for _, f := range files {
		got = append(got, f.Name())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected %v files, got %v file(s)", want, got)
	}
}

func TestFileSink_Reopen(t *testing.T) {
	tests := map[string]struct {
		Path string
	}{
		"stdout": {

			Path: stdout,
		},
		"stderr": {

			Path: stderr,
		},
		"dev/null": {

			Path: devnull,
		},
		"default-file": {},
	}

	isSpecialPath := func(path string) bool {
		switch path {
		case stdout, stderr, devnull:
			return true
		default:
			return false
		}
	}

	for name, tc := range tests {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			isSpecial := isSpecialPath(tc.Path)

			var path string
			switch {
			case isSpecial:
				// Use the path 'as is' since it will be a special type
				path = tc.Path
			default:
				path = t.TempDir()
			}

			fs := FileSink{
				Path:     path,
				FileName: "audit.log",
			}

			event := &Event{
				Formatted: map[string][]byte{JSONFormat: []byte("first")},
				Payload:   "First entry",
			}

			_, err := fs.Process(context.Background(), event)
			require.NoError(t, err)

			if !isSpecial {
				// delete file
				err = os.Remove(filepath.Join(path, "audit.log"))
				require.NoError(t, err)
			}

			// reopen
			err = fs.Reopen()
			require.NoError(t, err)

			event = &Event{
				Formatted: map[string][]byte{JSONFormat: []byte("second")},
				Payload:   "Second entry",
			}

			_, err = fs.Process(context.Background(), event)
			require.NoError(t, err)

			if !isSpecial {
				// Ensure process re-created the file
				dat, err := os.ReadFile(filepath.Join(path, "audit.log"))
				require.NoError(t, err)

				got := string(dat)
				want := "second"
				if got != "second" {
					t.Errorf("Expected file content to be %s, got %s", want, got)
				}

				files := 1
				if got, _ := os.ReadDir(path); len(got) != files {
					t.Errorf("Expected %d files, got %v file(s)", files, len(got))
				}
			}
		})
	}
}

func TestFileSink_TimeRotate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 2 * time.Second,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := 2
	if got, _ := os.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_TimestampOnlyOnRotate_TimeRotate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:                  tmpDir,
		FileName:              "audit.log",
		MaxDuration:           2 * time.Second,
		TimestampOnlyOnRotate: true,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("First entry")},
		Payload:   "First entry",
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("Last entry")},
		Payload:   "Last entry",
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := 2
	got, _ := os.ReadDir(tmpDir)
	if len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
	if got[1].Name() != "audit.log" {
		t.Errorf("Expected audit.log but found: %q", got[1].Name())
	}
	contents, _ := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	if expected := []byte("Last entry"); !bytes.Equal(contents, expected) {
		t.Errorf("Expected %q but found %q", string(expected), string(contents))
	}
}

func TestFileSink_ByteRotate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxBytes:    5,
		MaxDuration: 24 * time.Hour,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("entry")},
		Payload:   "entry",
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("entry")},
		Payload:   "entry",
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := 2
	if got, _ := os.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_TimestampOnlyOnRotate_ByteRotate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:                  tmpDir,
		FileName:              "audit.log",
		MaxBytes:              5,
		MaxDuration:           24 * time.Hour,
		TimestampOnlyOnRotate: true,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first entry")},
		Payload:   "first entry",
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("last entry")},
		Payload:   "last entry",
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := 2
	got, _ := os.ReadDir(tmpDir)
	if len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
	if got[1].Name() != "audit.log" {
		t.Errorf("Expected audit.log but found: %q", got[1].Name())
	}
	contents, _ := os.ReadFile(filepath.Join(tmpDir, "audit.log"))
	if expected := []byte("last entry"); !bytes.Equal(contents, expected) {
		t.Errorf("Expected %q but found %q", string(expected), string(contents))
	}
}

func TestFileSink_open(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Second,
	}
	err := fs.open()
	require.NoError(t, err)

	_, err = os.ReadFile(fs.f.Name())
	require.NoError(t, err)
}

func TestFileSink_pruneFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Hour,
		MaxBytes:    10,
		MaxFiles:    1,
	}

	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first entry")},
	}
	_, err := fs.Process(context.Background(), event)
	require.NoError(t, err)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("second entry")},
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("third entry")},
	}
	_, err = fs.Process(context.Background(), event)
	require.NoError(t, err)

	want := 2
	tmpFiles, _ := os.ReadDir(tmpDir)
	got := len(tmpFiles)
	if want != got {
		t.Errorf("Expected %d files, got %d", want, got)
	}
}
func TestFileSink_FileMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configuredFileMode := os.FileMode(0640)
	fs := FileSink{
		Path:     tmpDir,
		FileName: "audit.log",
		Mode:     configuredFileMode,
	}
	err := fs.open()
	require.NoError(t, err)

	fileInfo, err := os.Stat(fs.f.Name())
	require.NoError(t, err)

	// Ensure the file mode matches the desired mode
	actualMode := fileInfo.Mode()
	if actualMode != configuredFileMode {
		t.Errorf("Expected file mode %q, got %q", configuredFileMode.Perm(), actualMode.Perm())
	}
}

func TestFileSink_DirMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	parentDirMode := os.FileMode(0750)

	// Change mode on parent directory
	err := os.Chmod(tmpDir, parentDirMode)
	require.NoError(t, err)

	fs := FileSink{
		Path:     tmpDir,
		FileName: "audit.log",
	}
	err = fs.open()
	require.NoError(t, err)

	dirInfo, err := os.Stat(tmpDir)
	require.NoError(t, err)

	// Ensure the parent directory's permissions remain unchanged
	actualDirMode := dirInfo.Mode()
	if actualDirMode.Perm() != parentDirMode.Perm() {
		t.Errorf("Expected file mode %q, got %q", parentDirMode.Perm(), actualDirMode.Perm())
	}
}
