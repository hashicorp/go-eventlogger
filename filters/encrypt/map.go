package encrypt

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

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
	// Pointer.  This is optional and the default operations (or overrides) will
	// apply when not specified
	Filter FilterOperation
}

// Taggable defines an interface for taggable maps
type Taggable interface {

	// Tags will return a set of pointer tags for the map
	Tags() ([]PointerTag, error)
}

type tMap struct {
	value          reflect.Value
	filtered       bool                // true when all fields have been filtered.
	filteredFields map[string]struct{} // not nil when only some fields have been filtered
}

type trackedMaps map[uintptr]*tMap

// newTrackedMaps will create a new trackedMaps
func newTrackedMaps(tm ...*tMap) (trackedMaps, error) {
	const op = "encrypt.(eventMaps).newTrackedMaps"
	maps := make(trackedMaps, len(tm))
	for i, m := range tm {
		if err := maps.trackMap(m); err != nil {
			return nil, fmt.Errorf("%s: new map parameter # %d is not a valid: %w", op, i, err)
		}
	}
	return maps, nil
}

// trackMap will add the map to the list of tracked maps
func (maps trackedMaps) trackMap(tm *tMap) error {
	const op = "encrypt.(eventMaps).trackMap"
	if tm == nil {
		return fmt.Errorf("%s: missing map: %w", op, ErrInvalidParameter)
	}

	tmPtr := tm.value.Pointer()
	tmKind := tm.value.Kind()

	var isMapPtr bool
	if tmKind == reflect.Ptr && tm.value.Elem().Kind() == reflect.Map {
		isMapPtr = true
	}
	switch {
	case isMapPtr || tmKind == reflect.Map || tm.value.Type() == reflect.TypeOf(&structpb.Struct{}):
		if tm.value.IsNil() {
			return fmt.Errorf("%s: map value is nil: %w", op, ErrInvalidParameter)
		}
		maps[tmPtr] = tm
		return nil
	default:
		return fmt.Errorf("%s: %s is not a valid parameter type: %w", op, tm.value.Type(), ErrInvalidParameter)
	}
}

// unfiltered returns all the maps which haven't been tracked as filtered
func (maps trackedMaps) unfiltered() []*tMap {
	unfiltered := make([]*tMap, 0, len(maps))
	for _, m := range maps {
		if m.filtered {
			continue
		}
		unfiltered = append(unfiltered, m)
	}
	return unfiltered
}

// processUnfiltered will process/filter all the maps being tracked which
// haven't been tracked as filtered and it will mark them as filtered.
func (maps trackedMaps) processUnfiltered(ctx context.Context, ef *Filter, filterOverrides map[DataClassification]FilterOperation, opt ...Option) error {
	const op = "encrypt.(eventMaps).processUnfiltered"
	if ef == nil {
		return fmt.Errorf("%s: missing filter node: %w", op, ErrInvalidParameter)
	}

	for _, m := range maps.unfiltered() {
		// we will mark the map as filtered at the bottom of this loop.
		var v reflect.Value
		switch {
		case m.value.Type() == reflect.TypeOf(&structpb.Struct{}):
			v = m.value.Elem().FieldByName("Fields")
		case m.value.Kind() == reflect.Ptr:
			v = m.value.Elem()
		default:
			v = m.value
		}
		if v.Kind() != reflect.Map {
			return fmt.Errorf("%s: unfiltered value (%s) is a not a map: %w", op, v.Kind(), ErrInvalidParameter)
		}

		classificationTag := &tagInfo{
			Classification: UnknownClassification,
			Operation:      UnknownOperation,
		}
		for _, key := range v.MapKeys() {
			if m.filteredFields != nil {
				if _, ok := m.filteredFields[key.String()]; ok {
					continue // already filtered
				}
			}
			field := v.MapIndex(key)
			ftype := field.Type()
			fkind := field.Kind()

			switch {
			// if the field is a string or []byte then we just need to sanitize it
			case ftype == reflect.TypeOf("") || ftype == reflect.TypeOf([]uint8{}):
				if err := ef.filterValue(ctx, field, classificationTag, opt...); err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			case ftype == reflect.TypeOf(wrapperspb.StringValue{}) || ftype == reflect.TypeOf(wrapperspb.BytesValue{}):
				if err := ef.filterValue(ctx, field.FieldByName("Value"), classificationTag, opt...); err != nil {
					return err
				}
			// if the field is a slice
			case fkind == reflect.Slice:
				switch {
				// if the field is a slice of string or slice of []byte
				case ftype == reflect.TypeOf([]string{}) || ftype == reflect.TypeOf([][]uint8{}):
					if err := ef.filterSlice(ctx, classificationTag, field, opt...); err != nil {
						return fmt.Errorf("%s: %w", op, err)
					}
				// if the field is a slice of structs, recurse through them...
				default:
					for i := 0; i < field.Len(); i++ {
						f := field.Index(i)
						if f.Kind() == reflect.Ptr {
							f = f.Elem()
						}
						if f.Kind() != reflect.Struct {
							continue
						}
						if err := ef.filterField(ctx, f, filterOverrides, opt...); err != nil {
							return fmt.Errorf("%s: %w", op, err)
						}
					}
				}
			// if the field is a struct
			case fkind == reflect.Struct:
				if err := ef.filterField(ctx, field, filterOverrides, opt...); err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

			case ftype == reflect.TypeOf(&structpb.Value{}):
				i := field.Interface().(*structpb.Value)
				switch {
				case reflect.TypeOf(i.Kind) == reflect.TypeOf(&structpb.Value_StringValue{}):
					s := i.GetStringValue()
					vv := reflect.Indirect(reflect.ValueOf(&s))
					if err := ef.filterValue(ctx, vv, &tagInfo{}, opt...); err != nil {
						return fmt.Errorf("%s: %w", op, err)
					}
					v.SetMapIndex(key, reflect.ValueOf(structpb.NewStringValue(vv.String())))
				}
			default:
				// intentional no-op
			}
		}
		// very important to mark the current map as filtered before iterating
		m.filtered = true
		m.filteredFields = nil
	}
	return nil
}
