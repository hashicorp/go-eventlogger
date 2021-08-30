package example

import "github.com/hashicorp/eventlogger/filters/encrypt"

// TestTaggedMap is a map that implements the Taggable interface for testing
type TestTaggedMap map[string]interface{}

// Tags implements the taggable interface for the TestTaggedMap type
func (t TestTaggedMap) Tags() ([]encrypt.PointerTag, error) {
	return []encrypt.PointerTag{
		{
			Pointer:        "", // want "empty pointerstructure pointer string"
			Classification: encrypt.SecretClassification,
			Filter:         encrypt.RedactOperation,
		},
		{
			Pointer:        "/foo",
			Classification: encrypt.DataClassification("secrt"), // want "invalid data classification: \"secrt\""
			Filter:         encrypt.RedactOperation,
		},
		{
			Pointer:        "/foo",
			Classification: encrypt.SecretClassification,
			Filter:         encrypt.FilterOperation("redct"), // want "invalid filter operation: \"redct\""
		},
	}, nil
}

type example struct {
	TaggedMap TestTaggedMap
}
