# encrypt package

The encrypt package implements a new Filter that supports filtering fields in an
event payload using a custom tag named `classified`.  This new tag supports two
fields. The first field tag is the classification of the data (valid values are
public, sensitive and secret).  The second field is an optional filter operation
to apply (valid values are redact, encrypt, hmac-sha256).

**tagged struct example**
```go
type testPayloadStruct struct {
    Public    string `classified:"public"`
    Sensitive string `classified:"sensitive,redact"` // example of classification,operation
    Secret    []byte `classified:"secret"`
}

```

encrypt.Filter supports filtering the following struct field types within an
event payload, when they are tagged with a `classified` tag:
* `string`
* `[]string`
* `[]byte`
* `[][]byte`

encrypt.Filter also supports filtering any field of type `map[string]interface{}` that implements the
`encrypt.Taggable` interface. 

The following DataClassifications are supported:
* PublicClassification
* SensitiveClassification
* SecretClassification

The following FilterOperations are supported:
* NoOperation: no filter operation is applied to the field data.
* RedactOperation: redact the field data. 
* EncryptOperation: encrypts the field data.
* HmacSha256Operation: HMAC sha-256 the field data.



# Taggable interface
`map[string]interface{}` fields in an event payloads can be filtered using a
`[]PointerTag` for the map. To be filtered a `map[string]interface{}` field is
required to implement a single function interface:
```go
// Taggable defines an interface for taggable maps
type Taggable interface {
	// Tags will return a set of pointer tags for the map
	Tags() ([]PointerTag, error)
}

// PointerTag provides the pointerstructure pointer string to get/set a key
// within a map[string]interface{} along with its DataClassification and
// FilterOperation.
type PointerTag struct {
	// Pointer is the pointerstructure pointer string to get/set a key within a
	// map[string]interface{}  See: https://github.com/mitchellh/pointerstructure
	Pointer string

	// Classification is the DataClassification of data pointed to by the
	// Pointer
	Classification DataClassification

	// Filter is the FilterOperation to apply to the data pointed to by the
	// Pointer
	Filter FilterOperation
}
``` 

# Filter operation overrides

The Filter node will contain an optional field:

`FilterOperationOverrides map[DataClassification]FilterOperation`

This map can provide an optional set of runtime overrides for the FilterOperations to be applied to DataClassifications.

Normally, the filter operation applied to a field is determined by the operation
specified in its classified tag. If no operation is specified in the tag, then a
set of reasonable default filter operations are applied. 

FilterOperationOverrides provides the ability to override an event's "classified" tag settings.


# Default filter operations
* PublicClassification: NoOperation
* SensitiveClassification: EncryptOperation
* SecretClassification: RedactOperation
* NoClassification: RedactOperation