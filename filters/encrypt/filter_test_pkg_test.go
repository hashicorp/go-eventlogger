package encrypt_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/eventlogger/filters/encrypt"
	"github.com/hashicorp/eventlogger/filters/encrypt/testing/resources/protopayload"
	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type testPayloadStruct struct {
	PublicId          string `class:"public"`
	SensitiveUserName string `class:"sensitive"`
	LoginTimestamp    time.Time
	TaggedMap         encrypt.TestTaggedMap
}

type testPayloadStructWithTaggableSlice struct {
	PublicId          string `class:"public"`
	SensitiveUserName string `class:"sensitive"`
	LoginTimestamp    time.Time
	TaggedMap         []*encrypt.TestTaggedMap
}

type testPayload struct {
	notExported        string
	NotTagged          string
	SensitiveRedacted  []byte `class:"sensitive,redact"`
	StructPtr          *testPayloadStruct
	StructValue        testPayloadStruct
	StructPtrSlice     []*testPayloadStruct
	StructValueSlice   []testPayloadStruct
	Keys               [][]byte `class:"secret"`
	NeverFiltered      *testIgnore
	NeverFilteredSlice []*testIgnore
}

type testIgnore struct {
	NeverFiltered string
}

type testWrapperPayload struct {
	wrapper wrapping.Wrapper
	salt    []byte
	info    []byte
}

func (t *testWrapperPayload) Wrapper() wrapping.Wrapper { return t.wrapper }
func (t *testWrapperPayload) HmacSalt() []byte          { return t.salt }
func (t *testWrapperPayload) HmacInfo() []byte          { return t.info }

type testEventWrapperPayload struct {
	eventId     string
	salt        []byte
	info        []byte
	StructValue testPayloadStruct
}

func (t *testEventWrapperPayload) EventId() string  { return t.eventId }
func (t *testEventWrapperPayload) HmacSalt() []byte { return t.salt }
func (t *testEventWrapperPayload) HmacInfo() []byte { return t.info }

type testNilInterface interface {
	IsNil() bool
}
type testNilInterfaceStruct struct{}

func (t *testNilInterfaceStruct) IsNil() bool { return true }

func TestFilter_Process(t *testing.T) {
	ctx := context.Background()
	wrapper := encrypt.TestWrapper(t)
	now := time.Now()
	testEncryptingFilter := &encrypt.Filter{
		Wrapper:     wrapper,
		HmacSalt:    []byte("salt"),
		HmacInfo:    []byte("info"),
		IgnoreTypes: []reflect.Type{reflect.TypeOf(&testIgnore{})},
	}

	testString := "test-string"

	var nilInterface testNilInterface
	var foobar *testNilInterfaceStruct
	nilInterface = foobar

	b, err := structpb.NewValue([]byte("Tagged-Bytes"))
	require.NoError(t, err)

	testStructpbStruct, err := structpb.NewStruct(map[string]interface{}{
		encrypt.TestMapField: "alice",
	})
	require.NoError(t, err)

	testTaggable := &protopayload.WithTaggable{
		PublicString:            "PublicString",
		SensitiveString:         "SensitiveString",
		SecretString:            "SecretString",
		UnclassifiedString:      "UnclassifiedString",
		PublicStringValue:       &wrapperspb.StringValue{Value: "PublicStringValue"},
		SensitiveStringValue:    &wrapperspb.StringValue{Value: "SensitiveStringValue"},
		SecretStringValue:       &wrapperspb.StringValue{Value: "SecretStringValue"},
		UnclassifiedStringValue: &wrapperspb.StringValue{Value: "UnclassifiedStringValue"},
		PublicBytes:             []byte("PublicBytes"),
		SensitiveBytes:          []byte("SensitiveBytes"),
		SecretBytes:             []byte("SecretBytes"),
		UnclassifiedBytes:       []byte("UnclassifiedBytes"),
		PublicBytesValue:        &wrapperspb.BytesValue{Value: []byte("PublicBytesValue")},
		SensitiveBytesValue:     &wrapperspb.BytesValue{Value: []byte("SensitiveBytesValue")},
		SecretBytesValue:        &wrapperspb.BytesValue{Value: []byte("SecretBytesValue")},
		UnclassifiedBytesValue:  &wrapperspb.BytesValue{Value: []byte("UnclassifiedBytesValue")},
		TaggableAttributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				protopayload.TaggedStringField:   structpb.NewStringValue("Tagged"),
				protopayload.UntaggedStringField: structpb.NewStringValue("Untagged"),
				protopayload.TaggedBytesField:    b,
				protopayload.IntField:            structpb.NewNumberValue(10),
			},
		},
		NontaggableAttributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				protopayload.UntaggedStringField: structpb.NewStringValue("Untagged"),
			},
		},
		EmbeddedTaggable: &protopayload.EmbeddedTaggable{
			EPublicString: "PublicString",
			ESecretString: "SecretString",
			ETaggableAttributes: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					protopayload.TaggedStringField:   structpb.NewStringValue("Tagged"),
					protopayload.UntaggedStringField: structpb.NewStringValue("Untagged"),
				},
			},
		},
	}

	tests := []struct {
		name            string
		filter          *encrypt.Filter
		testEvent       *eventlogger.Event
		setupWantEvent  func(*eventlogger.Event)
		wantEvent       *eventlogger.Event
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "all-operations-are-no-op",
			filter: &encrypt.Filter{
				FilterOperationOverrides: func() map[encrypt.DataClassification]encrypt.FilterOperation {
					ops := encrypt.DefaultFilterOperations()
					for k := range ops {
						ops[k] = encrypt.NoOperation
					}
					return ops
				}(),
			},
			// this payload would normally raise an error because it needs
			// redacting and the string is not settable... but since we're
			// overriding all the filter operations to be NoOperation, it will
			// succeed and it's a simple way to test if the no-op short
			// circuiting is working when all the filter operations for a node
			// are set to NoOperation
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   testString,
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   testString,
			},
		},
		{
			name:   "simple",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					notExported:       "not-exported",
					NotTagged:         "not-tagged-data-will-be-redacted",
					SensitiveRedacted: []byte("sensitive-redact-override"),
					StructPtr: &testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
					StructValue: testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
					StructPtrSlice: []*testPayloadStruct{
						{
							PublicId:          "id-12",
							SensitiveUserName: "Alice Eve Doe",
						},
					},
					StructValueSlice: []testPayloadStruct{
						{
							PublicId:          "id-12",
							SensitiveUserName: "Alice Eve Doe",
						},
					},
					Keys: [][]byte{[]byte("key1"), []byte("key2")},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					NotTagged:         encrypt.RedactedData,
					SensitiveRedacted: []byte(encrypt.RedactedData),
					StructPtr: &testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe", // this will be decryped by the setupWantEvent func before comparison
					},
					StructValue: testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe", // this will be decryped by the setupWantEvent func before comparison
					},
					StructPtrSlice: []*testPayloadStruct{
						{
							PublicId:          "id-12",
							SensitiveUserName: "Alice Eve Doe",
						},
					},
					StructValueSlice: []testPayloadStruct{
						{
							PublicId:          "id-12",
							SensitiveUserName: "Alice Eve Doe",
						},
					},
					Keys: [][]byte{[]byte(encrypt.RedactedData), []byte(encrypt.RedactedData)},
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.(*testPayload).StructPtr.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructPtr.SensitiveUserName)))
				e.Payload.(*testPayload).StructValue.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructValue.SensitiveUserName)))
				e.Payload.(*testPayload).StructPtrSlice[0].SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructPtrSlice[0].SensitiveUserName)))
				e.Payload.(*testPayload).StructValueSlice[0].SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructValueSlice[0].SensitiveUserName)))
			},
		},
		{
			name:   "nil-byte-fields",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					NotTagged:         "not-tagged-data-will-be-redacted",
					SensitiveRedacted: nil,
					StructPtr: &testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
					Keys: nil,
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					NotTagged:         encrypt.RedactedData,
					SensitiveRedacted: nil,
					StructPtr: &testPayloadStruct{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe", // this will be decryped by the setupWantEvent func before comparison
					},
					Keys: nil,
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.(*testPayload).StructPtr.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructPtr.SensitiveUserName)))
				e.Payload.(*testPayload).StructValue.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructValue.SensitiveUserName)))
			},
		},
		{
			name:   "taggable",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &encrypt.TestTaggedMap{
					encrypt.TestMapField: "bar",
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &encrypt.TestTaggedMap{
					encrypt.TestMapField: encrypt.RedactedData,
				},
			},
		},
		{
			name:   "taggable-value",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: encrypt.TestTaggedMap{
					encrypt.TestMapField: "bar",
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: encrypt.TestTaggedMap{
					encrypt.TestMapField: encrypt.RedactedData,
				},
			},
		},
		{
			name:   "struct-with-taggable",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
					TaggedMap: encrypt.TestTaggedMap{
						encrypt.TestMapField: "bar",
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
					TaggedMap: encrypt.TestTaggedMap{
						encrypt.TestMapField: encrypt.RedactedData,
					},
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.(*testPayloadStruct).SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayloadStruct).SensitiveUserName)))
			},
		},
		{
			name:   "struct-with-slice-of-taggable",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayloadStructWithTaggableSlice{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
					TaggedMap: []*encrypt.TestTaggedMap{
						{
							encrypt.TestMapField:       "bar",
							encrypt.TestPublicMapField: "public-bar",
						},
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayloadStructWithTaggableSlice{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
					TaggedMap: []*encrypt.TestTaggedMap{
						{
							encrypt.TestMapField:       encrypt.RedactedData,
							encrypt.TestPublicMapField: "public-bar",
						},
					},
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				p := e.Payload.(*testPayloadStructWithTaggableSlice)
				p.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(p.SensitiveUserName)))
			},
		},
		{
			name:   "slice-of-taggable",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []encrypt.TestTaggedMap{
					{
						encrypt.TestMapField:       "bar",
						encrypt.TestPublicMapField: "public-bar",
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []encrypt.TestTaggedMap{
					{
						encrypt.TestMapField:       encrypt.RedactedData,
						encrypt.TestPublicMapField: "public-bar",
					},
				},
			},
		},
		{
			name:   "proto-with-taggable",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   testTaggable,
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					dup, err := copystructure.Copy(testTaggable)
					require.NoError(t, err)
					taggable := dup.(*protopayload.WithTaggable)

					taggable.SecretString = encrypt.RedactedData
					taggable.UnclassifiedString = encrypt.RedactedData
					taggable.SecretStringValue.Value = encrypt.RedactedData
					taggable.UnclassifiedStringValue.Value = encrypt.RedactedData

					taggable.SecretBytes = []byte(encrypt.RedactedData)
					taggable.UnclassifiedBytes = []byte(encrypt.RedactedData)
					taggable.SecretBytesValue.Value = []byte(encrypt.RedactedData)
					taggable.UnclassifiedBytesValue.Value = []byte(encrypt.RedactedData)
					taggable.TaggableAttributes.Fields[protopayload.UntaggedStringField] = structpb.NewStringValue(encrypt.RedactedData)

					taggable.TaggableAttributes.Fields[protopayload.TaggedStringField] = structpb.NewStringValue(encrypt.RedactedData) // overridden by Tags()
					taggable.NontaggableAttributes.Fields[protopayload.UntaggedStringField] = structpb.NewStringValue(encrypt.RedactedData)
					taggable.EmbeddedTaggable.ESecretString = encrypt.RedactedData
					taggable.EmbeddedTaggable.ETaggableAttributes.Fields[protopayload.UntaggedStringField] = structpb.NewStringValue(encrypt.RedactedData)

					return taggable
				}(),
			},
			setupWantEvent: func(e *eventlogger.Event) {
				taggable := e.Payload.(*protopayload.WithTaggable)
				taggable.SensitiveString = string(encrypt.TestDecryptValue(t, wrapper, []byte(taggable.SensitiveString)))
				taggable.SensitiveStringValue = &wrapperspb.StringValue{Value: string(encrypt.TestDecryptValue(t, wrapper, []byte(taggable.SensitiveStringValue.Value)))}

				taggable.SensitiveBytes = encrypt.TestDecryptValue(t, wrapper, taggable.SensitiveBytes)
				taggable.SensitiveBytesValue = &wrapperspb.BytesValue{Value: encrypt.TestDecryptValue(t, wrapper, taggable.SensitiveBytesValue.Value)}

				s := taggable.EmbeddedTaggable.ETaggableAttributes.Fields[protopayload.TaggedStringField].GetStringValue()
				taggable.EmbeddedTaggable.ETaggableAttributes.Fields[protopayload.TaggedStringField] = structpb.NewStringValue(string(encrypt.TestDecryptValue(t, wrapper, []byte(s))))

				s = taggable.TaggableAttributes.Fields[protopayload.TaggedBytesField].GetStringValue()
				taggable.TaggableAttributes.Fields[protopayload.TaggedBytesField] = structpb.NewStringValue(string(encrypt.TestDecryptValue(t, wrapper, []byte(s))))
				b, _ := base64.StdEncoding.DecodeString(taggable.TaggableAttributes.Fields[protopayload.TaggedBytesField].GetStringValue())
				t.Log(fmt.Sprintf("decoded TaggableAttributes[%s]", protopayload.TaggedBytesField), "==", string(b))
			},
		},
		{
			name:   "nil-interface",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: struct {
					Int    int
					T      testNilInterface
					SliceT []testNilInterface
				}{
					T:      nilInterface,
					SliceT: []testNilInterface{nilInterface},
					Int:    1,
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: struct {
					Int    int
					T      testNilInterface
					SliceT []testNilInterface
				}{
					T:      nil,
					SliceT: []testNilInterface{nil},
					Int:    1,
				},
			},
		},
		{
			name:   "nil-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   nil,
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   nil,
			},
		},
		{
			name:   "nil-ptr",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   func() interface{} { var p *testPayloadStruct; return p }(),
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   nil,
			},
		},
		{
			name:   "nil-interface",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					var taggable encrypt.Taggable
					return taggable
				}(),
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   nil,
			},
		},
		{
			name:   "slice-struct-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []testPayloadStruct{
					{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []testPayloadStruct{
					{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.([]testPayloadStruct)[0].SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.([]testPayloadStruct)[0].SensitiveUserName)))
			},
		},
		{
			name:   "slice-struct-ptr-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []*testPayloadStruct{
					{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []*testPayloadStruct{
					{
						PublicId:          "id-12",
						SensitiveUserName: "Alice Eve Doe",
					},
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.([]*testPayloadStruct)[0].SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.([]*testPayloadStruct)[0].SensitiveUserName)))
			},
		},
		{
			name:   "ptr-slice-string-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   &[]string{"test"},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   &[]string{encrypt.RedactedData},
			},
		},
		{
			name:   "slice-string-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   []string{"test"},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   []string{encrypt.RedactedData},
			},
		},
		{
			name:   "ptr-slice-string-ptr-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					s := "test"
					return &[]*string{&s}
				}(),
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					s := encrypt.RedactedData
					return &[]*string{&s}
				}(),
			},
		},
		{
			name:   "slice-string-ptr-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					s := "test"
					return []*string{&s}
				}(),
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					s := encrypt.RedactedData
					return []*string{&s}
				}(),
			},
		},
		{
			name:   "slice-map-structpbstruct",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					return []*structpb.Struct{
						testStructpbStruct,
					}
				}(),
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: func() interface{} {
					tmp, err := structpb.NewStruct(map[string]interface{}{
						encrypt.TestMapField: encrypt.RedactedData,
					})
					require.NoError(t, err)
					return []*structpb.Struct{tmp}
				}(),
			},
		},
		{
			name:   "string-ptr-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   &testString,
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   func() interface{} { s := encrypt.RedactedData; return &s }(),
			},
		},
		{
			name:   "string-payload",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload:   testString,
			},
			wantErrIs:       encrypt.ErrInvalidParameter,
			wantErrContains: "unable to redact string payload (not setable)",
		},
		{
			name:            "missing-event",
			filter:          testEncryptingFilter,
			testEvent:       nil,
			wantErrIs:       encrypt.ErrInvalidParameter,
			wantErrContains: "missing event",
		},
		{
			name:   "missing-wrapper",
			filter: &encrypt.Filter{},
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
				},
			},
			wantErrIs:       encrypt.ErrInvalidParameter,
			wantErrContains: "missing wrapper",
		},
		{
			name: "missing-wrapper-but-only-redacting",
			filter: &encrypt.Filter{
				FilterOperationOverrides: map[encrypt.DataClassification]encrypt.FilterOperation{
					encrypt.SecretClassification:    encrypt.RedactOperation,
					encrypt.SensitiveClassification: encrypt.RedactOperation,
				},
			},
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					SensitiveRedacted: []byte("sensitive-redacted"),
					NotTagged:         "not-tagged",
					StructValue: testPayloadStruct{
						SensitiveUserName: "sensitive-username",
					},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					SensitiveRedacted: []byte(encrypt.RedactedData),
					NotTagged:         encrypt.RedactedData,
					StructValue: testPayloadStruct{
						SensitiveUserName: encrypt.RedactedData,
					},
				},
			},
		},
		{
			name:   "with-ignore-type",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testIgnore{
					NeverFiltered: "never-filtered-value",
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testIgnore{
					NeverFiltered: "never-filtered-value",
				},
			},
		},
		{
			name:   "with-ignore-type-slice",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []*testIgnore{
					{NeverFiltered: "never-filtered-value"},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: []*testIgnore{
					{NeverFiltered: "never-filtered-value"},
				},
			},
		},
		{
			name:   "structwith-ignore-type",
			filter: testEncryptingFilter,
			testEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					NeverFiltered:      &testIgnore{NeverFiltered: "never-filtered-value"},
					NeverFilteredSlice: []*testIgnore{{NeverFiltered: "never-filtered-value"}},
				},
			},
			wantEvent: &eventlogger.Event{
				Type:      "test",
				CreatedAt: now,
				Payload: &testPayload{
					NeverFiltered:      &testIgnore{NeverFiltered: "never-filtered-value"},
					NeverFilteredSlice: []*testIgnore{{NeverFiltered: "never-filtered-value"}},
					NotTagged:          encrypt.RedactedData,
				},
			},
			setupWantEvent: func(e *eventlogger.Event) {
				e.Payload.(*testPayload).StructValue.SensitiveUserName = string(encrypt.TestDecryptValue(t, wrapper, []byte(e.Payload.(*testPayload).StructValue.SensitiveUserName)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)

			got, err := tt.filter.Process(ctx, tt.testEvent)
			if tt.wantErrIs != nil {
				require.Error(err)
				assert.ErrorIs(err, tt.wantErrIs)
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			actualJson, err := json.Marshal(got)
			require.NoError(err)
			t.Log(string(actualJson))

			if tt.setupWantEvent != nil {
				tt.setupWantEvent(got)
			}

			actualJson, err = json.Marshal(got)
			require.NoError(err)
			t.Log(string(actualJson))

			wantJson, err := json.Marshal(tt.wantEvent)
			require.NoError(err)
			assert.JSONEq(string(wantJson), string(actualJson))
		})
	}
	t.Run("rotate-wrapper-payload", func(t *testing.T) {
		t.Parallel()
		assert, require := assert.New(t), require.New(t)
		wrapper := encrypt.TestWrapper(t)
		ef := &encrypt.Filter{
			Wrapper:  wrapper,
			HmacSalt: []byte("salt"),
			HmacInfo: []byte("info"),
		}
		rotatedWrapper := encrypt.TestWrapper(t)
		e := &eventlogger.Event{
			Payload: &testWrapperPayload{
				wrapper: rotatedWrapper,
				info:    []byte("rotated-info"),
				salt:    []byte("rotated-salt"),
			},
		}
		got, err := ef.Process(context.Background(), e)
		require.NoError(err)
		assert.Nil(got)
		assert.Equal(rotatedWrapper, ef.Wrapper)
		assert.Equal([]byte("rotated-info"), ef.HmacInfo)
		assert.Equal([]byte("rotated-salt"), ef.HmacSalt)
	})
	t.Run("event-wrapper-info-payload-encrypt", func(t *testing.T) {
		t.Parallel()
		assert, require := assert.New(t), require.New(t)
		wrapper := encrypt.TestWrapper(t)
		ef := &encrypt.Filter{
			Wrapper:  wrapper,
			HmacSalt: []byte("salt"),
			HmacInfo: []byte("info"),
		}
		now := time.Now()
		e := &eventlogger.Event{
			Type:      "test",
			CreatedAt: now,
			Payload: &testEventWrapperPayload{
				eventId: "event-id",
				info:    []byte("event-info"),
				salt:    []byte("event-salt"),
				StructValue: testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
				},
			},
		}
		want := &eventlogger.Event{
			Type:      "test",
			CreatedAt: now,
			Payload: &testEventWrapperPayload{
				StructValue: testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
				},
			},
		}
		got, err := ef.Process(context.Background(), e)
		require.NoError(err)
		assert.NotNil(got)
		assert.Equal(wrapper, ef.Wrapper)
		assert.Equal([]byte("info"), ef.HmacInfo)
		assert.Equal([]byte("salt"), ef.HmacSalt)

		// assert that the node made a copy of the event before modifying it.
		assert.NotEqual(e.Payload.(*testEventWrapperPayload).StructValue.SensitiveUserName, got.Payload.(*testEventWrapperPayload).StructValue.SensitiveUserName)

		eventWrapper, err := encrypt.NewEventWrapper(wrapper, "event-id")
		require.NoError(err)
		got.Payload.(*testEventWrapperPayload).StructValue.SensitiveUserName = string(encrypt.TestDecryptValue(t, eventWrapper, []byte(got.Payload.(*testEventWrapperPayload).StructValue.SensitiveUserName)))
		assert.Equal(want, got)
	})
	t.Run("event-wrapper-info-payload-hmac", func(t *testing.T) {
		t.Parallel()
		assert, require := assert.New(t), require.New(t)
		now := time.Now()
		wrapper := encrypt.TestWrapper(t)
		ef := &encrypt.Filter{
			Wrapper:  wrapper,
			HmacSalt: []byte("salt"),
			HmacInfo: []byte("info"),
		}
		e := &eventlogger.Event{
			Type:      "test",
			CreatedAt: now,
			Payload: &testEventWrapperPayload{
				eventId: "event-id",
				info:    []byte("event-info"),
				salt:    []byte("event-salt"),
				StructValue: testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
				},
			},
		}
		wantHmac := &eventlogger.Event{
			Type:      "test",
			CreatedAt: now,
			Payload: &testEventWrapperPayload{
				StructValue: testPayloadStruct{
					PublicId:          "id-12",
					SensitiveUserName: "Alice Eve Doe",
				},
			},
		}
		ef.FilterOperationOverrides = map[encrypt.DataClassification]encrypt.FilterOperation{
			encrypt.SensitiveClassification: encrypt.HmacSha256Operation,
		}

		got, err := ef.Process(context.Background(), e)
		require.NoError(err)
		assert.NotNil(got)
		testWrapper, err := encrypt.NewEventWrapper(wrapper, "event-id")
		require.NoError(err)
		wantHmac.Payload.(*testEventWrapperPayload).StructValue.SensitiveUserName = encrypt.TestHmacSha256(t, []byte("Alice Eve Doe"), testWrapper, []byte("event-salt"), []byte("event-info"))
		assert.Equal(wantHmac, got)
	})
}

func TestFilter_Type(t *testing.T) {
	t.Parallel()
	ef := &encrypt.Filter{}
	assert.Equalf(t, eventlogger.NodeTypeFilter, ef.Type(), "Type() should always return %s", eventlogger.NodeTypeFilter)
}

func TestFilter_Reopen(t *testing.T) {
	t.Parallel()
	ef := &encrypt.Filter{}
	require.NoErrorf(t, ef.Reopen(), "Reopen is a no op and should never return an error")
}

func TestFilter_Rotate(t *testing.T) {
	t.Parallel()

	initialWrapper := encrypt.TestWrapper(t)

	rotatedWrapper := encrypt.TestWrapper(t)

	tests := []struct {
		name         string
		node         *encrypt.Filter
		opt          []encrypt.Option
		wantWrapper  wrapping.Wrapper
		wantSalt     []byte
		wantwithInfo []byte
	}{
		{
			name: "wrapper-only",
			node: &encrypt.Filter{
				Wrapper:  initialWrapper,
				HmacSalt: []byte("initial-salt"),
				HmacInfo: []byte("initial-info"),
			},
			opt:          []encrypt.Option{encrypt.WithWrapper(rotatedWrapper)},
			wantWrapper:  rotatedWrapper,
			wantSalt:     []byte("initial-salt"),
			wantwithInfo: []byte("initial-info"),
		},
		{
			name: "salt-only",
			node: &encrypt.Filter{
				Wrapper:  initialWrapper,
				HmacSalt: []byte("initial-salt"),
				HmacInfo: []byte("initial-info"),
			},
			opt:          []encrypt.Option{encrypt.WithSalt([]byte("rotated-salt"))},
			wantWrapper:  initialWrapper,
			wantSalt:     []byte("rotated-salt"),
			wantwithInfo: []byte("initial-info"),
		},
		{
			name: "info-only",
			node: &encrypt.Filter{
				Wrapper:  initialWrapper,
				HmacSalt: []byte("initial-salt"),
				HmacInfo: []byte("initial-info"),
			},
			opt:          []encrypt.Option{encrypt.WithInfo([]byte("rotated-info"))},
			wantWrapper:  initialWrapper,
			wantSalt:     []byte("initial-salt"),
			wantwithInfo: []byte("rotated-info"),
		},
		{
			name: "rotate-everything",
			node: &encrypt.Filter{
				Wrapper:  initialWrapper,
				HmacSalt: []byte("initial-salt"),
				HmacInfo: []byte("initial-info"),
			},
			opt: []encrypt.Option{
				encrypt.WithWrapper(rotatedWrapper),
				encrypt.WithSalt([]byte("rotated-salt")),
				encrypt.WithInfo([]byte("rotated-info")),
			},
			wantWrapper:  rotatedWrapper,
			wantSalt:     []byte("rotated-salt"),
			wantwithInfo: []byte("rotated-info"),
		},
	}
	for _, tt := range tests {
		assert := assert.New(t)
		tt.node.Rotate(tt.opt...)
		assert.Equal(tt.wantWrapper, tt.node.Wrapper)
		assert.Equal(tt.wantSalt, tt.node.HmacSalt)
		assert.Equal(tt.wantwithInfo, tt.node.HmacInfo)
	}
}
