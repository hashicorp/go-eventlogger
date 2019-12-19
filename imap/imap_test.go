package imap

import (
	"testing"

	"github.com/go-test/deep"
)

func testEq(t *testing.T, a, b interface{}) {
	if diff := deep.Equal(a, b); diff != nil {
		t.Fatal(diff)
	}
}

func TestIMap(t *testing.T) {

	m1 := NewIMap(map[string]interface{}{
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
	})

	type test struct {
		m    *IMap
		path *Path
		val  interface{}
		ok   bool
	}

	tests := []test{
		{m1, NewPath("Name"), "Frodo", true},
		{m1, NewPath("Age"), 50, true},
		{m1, NewPath("Attributes", "Strength"), 8, true},
		{m1, NewPath("Attributes", "Constitution"), 18, true},
		{m1, NewPath("Attributes", "Wisdom"), nil, false},
		{m1, NewPath("Missing"), nil, false},
		{m1, NewPath("Also", "Missing"), nil, false},
	}

	m2 := m1.Set(NewPath("Age"), 51)
	tests = append(tests, []test{
		{m2, NewPath("Name"), "Frodo", true},
		{m2, NewPath("Age"), 51, true},
		{m2, NewPath("Attributes", "Strength"), 8, true},
		{m2, NewPath("Attributes", "Constitution"), 18, true},
		{m2, NewPath("Attributes", "Wisdom"), nil, false},
		{m2, NewPath("Missing"), nil, false},
		{m2, NewPath("Also", "Missing"), nil, false},
	}...)

	m3 := m2.Set(NewPath("Attributes", "Constitution"), 3)
	tests = append(tests, []test{
		{m3, NewPath("Name"), "Frodo", true},
		{m3, NewPath("Age"), 51, true},
		{m3, NewPath("Attributes", "Strength"), 8, true},
		{m3, NewPath("Attributes", "Constitution"), 3, true},
		{m3, NewPath("Attributes", "Wisdom"), nil, false},
		{m3, NewPath("Missing"), nil, false},
		{m3, NewPath("Also", "Missing"), nil, false},
	}...)

	m4 := m3.Set(NewPath("Attributes", "Wisdom"), 12)
	tests = append(tests, []test{
		{m4, NewPath("Name"), "Frodo", true},
		{m4, NewPath("Age"), 51, true},
		{m4, NewPath("Attributes", "Strength"), 8, true},
		{m4, NewPath("Attributes", "Constitution"), 3, true},
		{m4, NewPath("Attributes", "Wisdom"), 12, true},
		{m4, NewPath("Missing"), nil, false},
		{m4, NewPath("Also", "Missing"), nil, false},
	}...)

	m5 := m4.Delete(NewPath("Attributes", "Strength"))
	tests = append(tests, []test{
		{m5, NewPath("Name"), "Frodo", true},
		{m5, NewPath("Age"), 51, true},
		{m5, NewPath("Attributes", "Strength"), nil, false},
		{m5, NewPath("Attributes", "Constitution"), 3, true},
		{m5, NewPath("Attributes", "Wisdom"), 12, true},
		{m5, NewPath("Missing"), nil, false},
		{m5, NewPath("Also", "Missing"), nil, false},
	}...)

	for _, tt := range tests {
		val, ok := tt.m.Get(tt.path)
		testEq(t, val, tt.val)
		testEq(t, ok, tt.ok)
	}
}
