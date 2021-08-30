package example

import (
	"github.com/hashicorp/eventlogger/filters/encrypt"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestTaggedStruct is a structpb.Struct that implements the Taggable interface for testing
type TestTaggedStruct structpb.Struct

// Tags implements the taggable interface for the TestTaggedStruct type
func (t TestTaggedStruct) Tags() ([]encrypt.PointerTag, error) {
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
	TaggedStruct  TestTaggedStruct
	WrappedString wrapperspb.StringValue `class:"publc"` // want "invalid data classification: \"publc\""
	WrappedBytes  wrapperspb.BytesValue  `class:"publc"` // want "invalid data classification: \"publc\""
	WrappedBool1  wrapperspb.BoolValue   `class:"public"`
	WrappedBool2  wrapperspb.BoolValue   `class:"secret"` // want "invalid data classification for non-filterable type"
}
