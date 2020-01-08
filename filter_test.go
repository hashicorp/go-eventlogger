package eventlogger

import (
	"testing"
)

func TestFilter(t *testing.T) {
	predicate := func(e *Event) (bool, error) {
		return true, nil
	}
	f := &Filter{Predicate: predicate}

	e := &Event{}
	_, err := f.Process(e)
	if err != nil {
		t.Fatal(err)
	}
}
