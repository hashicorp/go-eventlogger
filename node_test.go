package eventlogger

import "testing"

func TestFilter(t *testing.T) {

	predicate := func(e *Event) (bool, error) {
		return true, nil
	}
	f := &Filter{Predicate: predicate, Next: nil}

	g := &Graph{}
	e := &Event{}
	err := f.Process(g, e)
	if err != nil {
		t.Fatal(err)
	}
}

func TestByteWriter(t *testing.T) {

	marshaller := func(e *Event) ([]byte, error) {
		return make([]byte, 0), nil
	}
	w := &ByteWriter{Marshaller: marshaller, Next: nil}

	g := &Graph{}
	e := &Event{}
	err := w.Process(g, e)
	if err != nil {
		t.Fatal(err)
	}
}
