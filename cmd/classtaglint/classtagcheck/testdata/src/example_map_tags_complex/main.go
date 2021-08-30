package example

import (
	"fmt"

	"github.com/hashicorp/eventlogger/filters/encrypt"
)

// TestTaggedMap is a map that implements the Taggable interface for testing
type TestTaggedMap map[string]interface{}

const Empty = ""

var FirstTag = encrypt.PointerTag{
	Pointer:        Empty, // want "empty pointerstructure pointer string"
	Classification: encrypt.SecretClassification,
	Filter:         encrypt.RedactOperation,
}

// Tags implements the taggable interface for the TestTaggedMap type
func (t TestTaggedMap) Tags() ([]encrypt.PointerTag, error) {
	var secondTag = encrypt.PointerTag{
		Pointer:        "",    // want "empty pointerstructure pointer string"
		Classification: Empty, // want "empty classification string"
		Filter:         encrypt.NoOperation,
	}

	thirdTag := encrypt.PointerTag{
		Pointer:        Empty,                               // want "empty pointerstructure pointer string"
		Classification: encrypt.DataClassification("secrt"), // want "invalid data classification: \"secrt\""
		Filter:         encrypt.RedactOperation,
	}

	fourthTag := encrypt.PointerTag{
		Pointer:        "/foo",
		Classification: encrypt.DataClassification("secrt"), // want "invalid data classification: \"secrt\""
		Filter:         encrypt.RedactOperation,
	}

	fourthTag.Pointer = Empty     // want "empty pointerstructure pointer string"
	fourthTag.Classification = "" // want "empty classification string"
	fourthTag.Filter = ""

	if fourthTag.Pointer == "" {
		return nil, fmt.Errorf("pointer tag is empty")
	}

	tags := []encrypt.PointerTag{
		FirstTag,
		secondTag,
		thirdTag,
		fourthTag,
		{
			func(value string) string { return value }(Empty), // analyzer will not try to resolve dynamic values
			encrypt.SecretClassification,
			encrypt.FilterOperation("redct"), // want "invalid filter operation: \"redct\""
		},
		{
			"" + "" + "", // want "empty pointerstructure pointer string"
			encrypt.SecretClassification,
			encrypt.FilterOperation("redct"), // want "invalid filter operation: \"redct\""
		},
		{
			"/foo",
			encrypt.SecretClassification,
			encrypt.FilterOperation("redct"), // want "invalid filter operation: \"redct\""
		},
	}

	tags = append(tags, thirdTag)

	return append(tags, fourthTag, encrypt.PointerTag{Empty, encrypt.PublicClassification, encrypt.NoOperation}), nil // want "empty pointerstructure pointer string"
}

type example struct {
	TaggedMap TestTaggedMap
}
