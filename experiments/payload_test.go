package experiments

import (
	"errors"
	"testing"

	"github.com/go-test/deep"
)

func assert(t *testing.T, a, b interface{}) {
	if diff := deep.Equal(a, b); diff != nil {
		t.Fatal(diff)
	}
}

func TestPayload(t *testing.T) {

	base := map[string]interface{}{
		"Name": "Frodo",
		"Age":  50,
		"Parents": []interface{}{
			"Drogo",
			"Primula",
		},
		"Attributes": map[string]interface{}{
			"Strength":     8,
			"Constitution": 18,
			"Dexterity":    12,
		},
	}

	p1 := NewPayload(base)

	type test struct {
		payload Payload
		path    *Path
		val     interface{}
		ok      bool
	}

	tests := []test{
		{p1, NewPath("Name"), "Frodo", true},
		{p1, NewPath("Age"), 50, true},
		{p1, NewPath("Attributes", "Strength"), 8, true},
		{p1, NewPath("Attributes", "Constitution"), 18, true},
	}

	p2, err := p1.Set(NewPath("Age"), 51)
	if err != nil {
		t.Fatal(err)
	}
	tests = append(tests, []test{
		{p2, NewPath("Name"), "Frodo", true},
		{p2, NewPath("Age"), 51, true},
		{p1, NewPath("Attributes", "Strength"), 8, true},
		{p2, NewPath("Attributes", "Constitution"), 18, true},
	}...)

	p3, err := p2.Set(NewPath("Attributes", "Constitution"), 3)
	if err != nil {
		t.Fatal(err)
	}
	tests = append(tests, []test{
		{p3, NewPath("Name"), "Frodo", true},
		{p3, NewPath("Age"), 51, true},
		{p1, NewPath("Attributes", "Strength"), 8, true},
		{p3, NewPath("Attributes", "Constitution"), 3, true},
	}...)

	for _, tt := range tests {
		val, ok := tt.payload.Get(tt.path)
		assert(t, val, tt.val)
		assert(t, ok, tt.ok)
	}

	_, err = p3.Set(NewPath("New", "Path"), 123)
	if err == nil {
		t.Fatal(errors.New("Expected error setting a value on a new path"))
	}
}
