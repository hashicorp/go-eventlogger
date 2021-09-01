package encrypt

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Test_newTrackedMaps(t *testing.T) {
	testTMap := reflect.ValueOf(map[string]string{"eve": "alice"})

	tests := []struct {
		name            string
		tm              []*tMap
		want            *trackedMaps
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "no-args",
			want: &trackedMaps{
				tracked: map[uintptr]*tMap{},
			},
		},
		{
			name: "bad-map-value",
			tm: []*tMap{
				{
					value: reflect.ValueOf(""),
				},
			},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "string is not a valid parameter type",
		},
		{
			name:            "nil-map",
			tm:              []*tMap{nil},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing map",
		},
		{
			name:            "nil-map-value",
			tm:              []*tMap{{}},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "map value is missing",
		},
		{
			name: "valid",
			tm: []*tMap{
				{
					value:    testTMap,
					filtered: true,
				},
			},
			want: &trackedMaps{
				tracked: map[uintptr]*tMap{
					testTMap.Pointer(): {
						value:    testTMap,
						filtered: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			got, err := newTrackedMaps(tt.tm...)
			if tt.wantErr {
				require.Error(err)
				assert.Nil(got)
				if tt.wantErrIs != nil {
					assert.ErrorIs(err, tt.wantErrIs)
				}
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			assert.Equal(tt.want, got)
		})
	}
}

func Test_trackedMaps_unfiltered(t *testing.T) {
	m1 := reflect.ValueOf(map[string]string{"bob": "eve"})
	m2 := reflect.ValueOf(map[string]string{"eve": "alice"})

	tests := []struct {
		name string
		tm   *trackedMaps
		want []*tMap
	}{
		{
			name: "simple",
			tm: &trackedMaps{
				tracked: map[uintptr]*tMap{
					m1.Pointer(): {value: m1, filtered: true},
					m2.Pointer(): {value: m2, filtered: false},
				},
			},
			want: []*tMap{
				{value: m2, filtered: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got := tt.tm.unfiltered()
			assert.Equal(tt.want, got)
		})
	}
}

func Test_trackedMaps_trackMap(t *testing.T) {
	testTMap := reflect.ValueOf(map[string]string{"eve": "alice"})

	tests := []struct {
		name            string
		tm              *trackedMaps
		m               *tMap
		wantTm          *trackedMaps
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:            "missing-map",
			tm:              &trackedMaps{},
			m:               nil,
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing map",
		},
		{
			name:            "missing-value",
			tm:              &trackedMaps{},
			m:               &tMap{},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "map value is missing",
		},
		{
			name: "not-valid-value",
			tm:   &trackedMaps{},
			m: &tMap{
				value: reflect.ValueOf(""),
			},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "string is not a valid parameter type",
		},
		{
			name: "valid",
			tm:   &trackedMaps{},
			m: &tMap{
				value: testTMap,
			},
			wantTm: &trackedMaps{
				tracked: map[uintptr]*tMap{
					testTMap.Pointer(): {
						value: testTMap,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			err := tt.tm.trackMap(tt.m)
			if tt.wantErr {
				require.Error(err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(err, tt.wantErrIs)
				}
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantTm, tt.tm)
		})
	}
}

func Test_trackedMaps_trackTaggable(t *testing.T) {
	// testTMap := reflect.ValueOf(map[string]string{"eve": "alice"})

	testMap := TestTaggedMap{
		TestMapField:       "alice",
		TestMapField + "2": "bob",
	}
	tests := []struct {
		name            string
		tm              *trackedMaps
		taggable        Taggable
		pointer         string
		wantTm          *trackedMaps
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:            "missing-pointer",
			tm:              &trackedMaps{},
			taggable:        TestTaggedMap{},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing pointer",
		},
		{
			name:            "missing-taggable",
			tm:              &trackedMaps{},
			pointer:         "/" + TestMapField,
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing taggable",
		},
		{
			name:            "bad-path",
			tm:              &trackedMaps{},
			pointer:         "missing-initial-path-delimiter",
			taggable:        TestTaggedMap{},
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "invalid taggable pointer",
		},
		{
			name:            "pointer-path-invalid",
			tm:              &trackedMaps{},
			pointer:         "/unknownMap/" + TestMapField,
			taggable:        TestTaggedMap{},
			wantErr:         true,
			wantErrContains: "/unknownMap at part 0: couldn't find key",
		},
		{
			name: "valid",
			tm: &trackedMaps{
				tracked: map[uintptr]*tMap{
					reflect.ValueOf(testMap).Pointer(): {
						value:          reflect.ValueOf(testMap),
						filteredFields: map[string]struct{}{},
					},
				},
			},
			pointer:  "/" + TestMapField,
			taggable: testMap,
			wantTm: &trackedMaps{
				tracked: map[uintptr]*tMap{
					reflect.ValueOf(testMap).Pointer(): {
						value: reflect.ValueOf(testMap),
						filteredFields: map[string]struct{}{
							TestMapField: {},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			err := tt.tm.trackTaggable(tt.taggable, tt.pointer)
			if tt.wantErr {
				require.Error(err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(err, tt.wantErrIs)
				}
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantTm, tt.tm)
		})
	}
}

func Test_trackedMaps_processUnfiltered(t *testing.T) {
	ctx := context.Background()
	ef := &Filter{}

	overrides := DefaultFilterOperations()
	for class, op := range overrides {
		switch op {
		case EncryptOperation, HmacSha256Operation:
			overrides[class] = RedactOperation
		}
	}

	newTm := func(m interface{}) *trackedMaps {
		return &trackedMaps{
			tracked: map[uintptr]*tMap{
				reflect.ValueOf(m).Pointer(): {
					value: reflect.ValueOf(m),
				},
			},
		}
	}

	testStructpbStruct, err := structpb.NewStruct(map[string]interface{}{
		TestMapField: "alice",
	})
	require.NoError(t, err)

	type testStruct struct {
		Name string
	}

	tests := []struct {
		name            string
		ef              *Filter
		tm              *trackedMaps
		wantErr         bool
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:            "missing-filter",
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "missing filter node",
		},
		{
			name:            "not-a-map",
			ef:              ef,
			tm:              newTm([]string{"string-slice"}),
			wantErr:         true,
			wantErrIs:       ErrInvalidParameter,
			wantErrContains: "unfiltered value (slice) is a not a map",
		},
		{
			name: "map-string-interface-with-nil-field",
			ef:   ef,
			tm:   newTm(map[string]interface{}{TestMapField: nil}),
		},
		{
			name: "map-string-ptr-with-nil",
			ef:   ef,
			tm:   newTm(map[string]*testStruct{TestMapField: nil}),
		},
		{
			name: "no-tracked-maps",
			ef:   ef,
			tm:   &trackedMaps{},
		},
		{
			name: "map-string-interface",
			ef:   ef,
			tm: newTm(map[string]interface{}{
				TestMapField: "alice",
			}),
		},
		{
			name: "map-string-interface-of-ptr-structpb",
			ef:   ef,
			tm: newTm(map[string]interface{}{
				TestMapField: structpb.NewStringValue("alice"),
			}),
		},
		{
			name: "map-string-string",
			ef:   ef,
			tm: newTm(map[string]string{
				TestMapField: "alice",
			}),
		},
		{
			name: "map-string-byte",
			ef:   ef,
			tm: newTm(map[string][]byte{
				TestMapField: []byte("alice"),
			}),
		},
		{
			name: "map-string-wrapperspb-string-value",
			ef:   ef,
			tm: newTm(map[string]wrapperspb.StringValue{
				TestMapField: {Value: "alice"},
			}),
		},
		{
			name: "map-string-wrapperspb-bytes-value",
			ef:   ef,
			tm: newTm(map[string]wrapperspb.BytesValue{
				TestMapField: {Value: []byte("alice")},
			}),
		},
		{
			name: "map-string-structpb-struct",
			ef:   ef,
			tm:   newTm(testStructpbStruct),
		},
		{
			name: "map-string-string-slice",
			ef:   ef,
			tm: newTm(map[string][]string{
				TestMapField: {"alice"},
			}),
		},
		{
			name: "map-string-string-bytes",
			ef:   ef,
			tm: newTm(map[string][]byte{
				TestMapField: []byte("alice"),
			}),
		},
		{
			name: "map-string-struct-name",
			ef:   ef,
			tm: newTm(map[string]*testStruct{
				TestMapField: {Name: "alice"},
			}),
		},
		{
			name: "map-string-slice-struct-name",
			ef:   ef,
			tm: newTm(map[string][]*testStruct{
				TestMapField: {
					{Name: "alice"},
				},
			}),
		},
		{
			name: "map-string-slice-map-string-string",
			ef:   ef,
			tm: newTm(map[string][]map[string]string{
				TestMapField: {
					{
						TestMapField: "alice",
					},
				},
			}),
		},
		{
			name: "map-string-slice-interface",
			ef:   ef,
			tm: newTm(map[string][]interface{}{
				TestMapField: {
					&testStruct{
						Name: "alice",
					},
				},
			}),
		},
		{
			name: "map-string-map-string-string",
			ef:   ef,
			tm: newTm(map[string]map[string]string{
				TestMapField: {
					TestMapField: "alice",
				},
			}),
		},
		{
			name: "map-string-map-structpbvalue",
			ef:   ef,
			tm: newTm(map[string]*structpb.Value{
				TestMapField: structpb.NewStringValue("alice"),
			}),
		},
		{
			name: "map-string-map-structpbstruct",
			ef:   ef,
			tm: newTm(map[string]*structpb.Struct{
				TestMapField: testStructpbStruct,
			}),
		},
		{
			name: "slice-map-structpbstruct",
			ef:   ef,
			tm: newTm(map[string][]*structpb.Struct{
				TestMapField: {testStructpbStruct},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			err := tt.tm.processUnfiltered(ctx, tt.ef, overrides)
			if tt.wantErr {
				require.Error(err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(err, tt.wantErrIs)
				}
				if tt.wantErrContains != "" {
					assert.Contains(err.Error(), tt.wantErrContains)
				}
				return
			}
			require.NoError(err)
			for _, m := range tt.tm.tracked {
				mValue := m.value
				if m.value.Type() == reflect.TypeOf(&structpb.Struct{}) {
					mValue = m.value.Elem().FieldByName("Fields")
				}
				for _, k := range mValue.MapKeys() {
					v := mValue.MapIndex(k)
					gotJson, err := json.Marshal(v.Interface())
					require.NoError(err)
					switch {
					case tt.name == "map-string-interface-with-nil-field" || tt.name == "map-string-ptr-with-nil":
						wantJson, err := json.Marshal(nil)
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf([]string{}):
						wantJson, err := json.Marshal([]string{RedactedData})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf([]uint8{}):
						assert.Equal(reflect.ValueOf([]byte(RedactedData)).Bytes(), v.Bytes())

					case v.Type() == reflect.TypeOf(wrapperspb.StringValue{}):
						wantJson, err := json.Marshal(wrapperspb.StringValue{Value: RedactedData})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf(wrapperspb.BytesValue{}):
						wantJson, err := json.Marshal(wrapperspb.BytesValue{Value: []byte(RedactedData)})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf(&testStruct{}):
						wantJson, err := json.Marshal(&testStruct{RedactedData})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf([]*testStruct{}):
						wantJson, err := json.Marshal([]*testStruct{{RedactedData}})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf([]map[string]string{}):
						wantJson, err := json.Marshal([]map[string]string{{TestMapField: RedactedData}})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf(map[string]string{}):
						wantJson, err := json.Marshal(map[string]string{TestMapField: RedactedData})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf(map[string]*structpb.Struct{}):
						wantJson, err := json.Marshal(
							map[string]*structpb.Value{
								TestMapField: structpb.NewStringValue(RedactedData),
							},
						)

						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf([]*structpb.Struct{}):
						s, err := structpb.NewStruct(map[string]interface{}{TestMapField: RedactedData})
						require.NoError(err)
						wantJson, err := json.Marshal([]*structpb.Struct{s})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case v.Type() == reflect.TypeOf(&structpb.Struct{}):
						s, err := structpb.NewStruct(map[string]interface{}{TestMapField: RedactedData})
						require.NoError(err)
						wantJson, err := json.Marshal(s)
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					case tt.name == "map-string-slice-interface":
						wantJson, err := json.Marshal([]interface{}{
							testStruct{Name: RedactedData},
						})
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))

					default:
						t.Log("default case for type...value: ", v.Type(), "...", v)
						wantJson, err := json.Marshal(RedactedData)
						require.NoError(err)
						assert.JSONEq(string(wantJson), string(gotJson))
					}
				}
			}
		})
	}
}
