package protopayload

import (
	"github.com/hashicorp/eventlogger/filters/encrypt"
)

const (
	TaggedStringField   = "tagged_string_field"
	UntaggedStringField = "not_tagged_string"
	TaggedBytesField    = "tagged_bytes_field"
)

// Tags satisfy the encrypt.Taggable interface and return the "tagged"
// fields/attributes within both the TaggableAttributes and
// EmbeddedTaggableAttributes fields.  This is used only for testing purposes.
func (pp *WithTaggable) Tags() ([]encrypt.PointerTag, error) {
	var ptrs []encrypt.PointerTag
	if pp.TaggableAttributes != nil {
		_, ok := pp.TaggableAttributes.GetFields()[TaggedStringField]
		if ok {
			ptrs = append(ptrs, encrypt.PointerTag{
				Pointer:        "/TaggableAttributes/Fields/" + TaggedStringField,
				Classification: encrypt.SensitiveClassification,
				Filter:         encrypt.RedactOperation,
			})
		}
		_, ok = pp.TaggableAttributes.GetFields()[TaggedBytesField]
		if ok {
			ptrs = append(ptrs, encrypt.PointerTag{
				Pointer:        "/TaggableAttributes/Fields/" + TaggedBytesField,
				Classification: encrypt.SensitiveClassification,
			})
		}
	}
	if pp.EmbeddedTaggable != nil && pp.EmbeddedTaggable.ETaggableAttributes != nil {
		_, ok := pp.EmbeddedTaggable.ETaggableAttributes.GetFields()[TaggedStringField]
		if ok {
			ptrs = append(ptrs, encrypt.PointerTag{
				Pointer:        "/EmbeddedTaggable/ETaggableAttributes/Fields/" + TaggedStringField,
				Classification: encrypt.SensitiveClassification,
			})
		}
	}
	return ptrs, nil
}
