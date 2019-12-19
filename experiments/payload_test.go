package experiments

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

func assert(t *testing.T, a, b interface{}) {
	if diff := deep.Equal(a, b); diff != nil {
		t.Fatal(diff)
	}
}

func TestPayload(t *testing.T) {

	p1 := NewPayload(map[string]interface{}{
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
		{p1, NewPath("Attributes", "Wisdom"), nil, false},
	}

	p2 := p1.Set(NewPath("Age"), 51)
	tests = append(tests, []test{
		{p2, NewPath("Name"), "Frodo", true},
		{p2, NewPath("Age"), 51, true},
		{p2, NewPath("Attributes", "Strength"), 8, true},
		{p2, NewPath("Attributes", "Constitution"), 18, true},
		{p2, NewPath("Attributes", "Wisdom"), nil, false},
	}...)

	p3 := p2.Set(NewPath("Attributes", "Constitution"), 3)
	tests = append(tests, []test{
		{p3, NewPath("Name"), "Frodo", true},
		{p3, NewPath("Age"), 51, true},
		{p3, NewPath("Attributes", "Strength"), 8, true},
		{p3, NewPath("Attributes", "Constitution"), 3, true},
		{p3, NewPath("Attributes", "Wisdom"), nil, false},
	}...)

	p4 := p3.Set(NewPath("Attributes", "Wisdom"), 12)
	tests = append(tests, []test{
		{p4, NewPath("Name"), "Frodo", true},
		{p4, NewPath("Age"), 51, true},
		{p4, NewPath("Attributes", "Strength"), 8, true},
		{p4, NewPath("Attributes", "Constitution"), 3, true},
		{p4, NewPath("Attributes", "Wisdom"), 12, true},
	}...)

	p5 := p4.Delete(NewPath("Attributes", "Strength"))
	tests = append(tests, []test{
		{p5, NewPath("Name"), "Frodo", true},
		{p5, NewPath("Age"), 51, true},
		{p5, NewPath("Attributes", "Strength"), nil, false},
		{p5, NewPath("Attributes", "Constitution"), 3, true},
		{p5, NewPath("Attributes", "Wisdom"), 12, true},
	}...)

	for i, tt := range tests {
		fmt.Printf("%d %s\n", i, tt.path)
		val, ok := tt.payload.Get(tt.path)
		assert(t, val, tt.val)
		assert(t, ok, tt.ok)
	}

}
