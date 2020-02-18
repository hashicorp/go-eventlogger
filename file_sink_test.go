package eventlogger

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFileSink_TimeRotate(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 2 * time.Second,
	}
	event := &Event{
		Formatted: map[string][]byte{"json": []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{"json": []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	if got, _ := ioutil.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_ByteRotate(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxBytes:    5,
		MaxDuration: 24 * time.Hour,
	}
	event := &Event{
		Formatted: map[string][]byte{"json": []byte("entry")},
		Payload:   "entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{"json": []byte("entry")},
		Payload:   "entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	if got, _ := ioutil.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_open(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Second,
	}
	err = fs.open()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ioutil.ReadFile(fs.f.Name())
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileSink_pruneFiles(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Hour,
		MaxBytes:    10,
		MaxFiles:    1,
	}

	event := &Event{
		Formatted: map[string][]byte{"json": []byte("first entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	event = &Event{
		Formatted: map[string][]byte{"json": []byte("second entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	event = &Event{
		Formatted: map[string][]byte{"json": []byte("third entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	tmpFiles, _ := ioutil.ReadDir(tmpDir)
	got := len(tmpFiles)
	if want != got {
		t.Errorf("Expected %d files, got %d", want, got)
	}

}
