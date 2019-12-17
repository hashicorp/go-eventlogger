package eventlogger

import "testing"

func TestFilter(t *testing.T) {

	predicate := func(e *Envelope) (bool, error) {
		return true, nil
	}
	f := &Filter{Predicate: predicate}

	e := &Envelope{}
	err := f.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestByteWriter(t *testing.T) {

	marshaller := func(e *Envelope) ([]byte, error) {
		return make([]byte, 0), nil
	}
	w := &ByteWriter{Marshaller: marshaller}

	e := &Envelope{}
	err := w.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileSink(t *testing.T) {
	fs := &FileSink{FilePath: "test.log"}
	e := &Envelope{}
	e.Marshalled = []byte("abcdef")
	err := fs.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}
