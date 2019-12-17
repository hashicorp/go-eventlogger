package eventlogger

import "testing"

func TestFilter(t *testing.T) {

	predicate := func(e *Event) (bool, error) {
		return true, nil
	}
	f := &Filter{Predicate: predicate}

	e := &Event{}
	err := f.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestByteWriter(t *testing.T) {

	marshaller := func(e *Event) ([]byte, error) {
		return make([]byte, 0), nil
	}
	w := &ByteWriter{Marshaller: marshaller}

	e := &Event{}
	err := w.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileSink(t *testing.T) {
	fs := &FileSink{FilePath: "test.log"}
	e := &Event{}
	e.Writable = []byte("abcdef")
	err := fs.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}
