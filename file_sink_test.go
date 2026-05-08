// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	t.Parallel()

	tests := map[string]struct {
		Path   string
		IsFile bool
	}{
		"stdout": {
			Path: FileStdout,
		},
		"stderr": {
			Path: FileStderr,
		},
		"dev/null": {
			Path: FileDevNull,
		},
		"default-file": {
			IsFile: true,
		},
	}

	for name, tc := range tests {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var path string
			switch {
			case tc.IsFile:
				// Create a temporary directory to store this file in.
				path = t.TempDir()
			default:
				// Use the path 'as is' since it will be a special type
				path = tc.Path
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

			if tc.IsFile {
				// manually delete the file if not a special path
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

			if tc.IsFile {
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

	time.Sleep(3 * time.Second)

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

	tests := map[string]struct {
		Path   string
		IsFile bool
	}{
		"stdout": {
			Path: "/dev/stdout",
		},
		"stderr": {
			Path: "/dev/stderr",
		},
		"null": {
			Path: "/dev/null",
		},
		"file": {
			Path:   t.TempDir(),
			IsFile: true,
		},
	}

	for name, tc := range tests {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := FileSink{
				Path:        tc.Path,
				FileName:    "audit.log",
				MaxDuration: 1 * time.Second,
			}
			err := fs.open()
			require.NoError(t, err)

			// If this path should have been for a real file, attempt to open
			// it from the operating system.
			if tc.IsFile {
				_, err = os.ReadFile(fs.f.Name())
				require.NoError(t, err)
			}
		})
	}
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

func TestFileSink_ContextCancellation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		path string
	}{
		"regular-file-path": {
			path: t.TempDir(),
		},
		"stdout": {
			path: FileStdout,
		},
		"stderr": {
			path: FileStderr,
		},
		"devnull": {
			path: FileDevNull,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := &FileSink{
				Path:     tc.path,
				FileName: "sink.log",
			}

			// Create and immediately cancel the context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			event := &Event{
				Formatted: map[string][]byte{JSONFormat: []byte(`{"msg":"test data"}`)},
				Payload:   "test",
			}

			// Process should return context error immediately.
			_, err := fs.Process(ctx, event)
			require.Error(t, err)
			require.Equal(t, context.Canceled, err)
		})
	}
}

func TestFileSink_ContextCancellationBetweenWrites(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fs := &FileSink{
		Path:     tmpDir,
		FileName: "sink.log",
	}

	// Create a context that we'll cancel between writes.
	ctx, cancel := context.WithCancel(context.Background())

	// Process one event successfully.
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte(`{"msg":"first event"}`)},
		Payload:   "first event",
	}
	_, err := fs.Process(ctx, event)
	require.NoError(t, err)

	// Verify the first event was processed and written to the sink.
	filePath := filepath.Join(tmpDir, "sink.log")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, `{"msg":"first event"}`, string(content))

	// Cancel the context and next time we process an event,
	// it should fail with context error.
	cancel()

	event2 := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte(`{"msg":"second event"}`)},
		Payload:   "second event",
	}
	_, err = fs.Process(ctx, event2)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)

	// Verify the file still only contains the first event (second write didn't happen).
	contentAfter, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, `{"msg":"first event"}`, string(contentAfter))
}
