package encrypt

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/pointerstructure"
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
	const op = "encrypt.(trackedMaps).newTrackedMaps"
	maps := make(trackedMaps, len(tm))
	for i, m := range tm {
		if err := maps.trackMap(m); err != nil {
			return nil, fmt.Errorf("%s: new map parameter #%d is not a valid: %w", op, i, err)
		}
	}
	return maps, nil
}

// trackMap will add the map to the list of tracked maps
func (maps trackedMaps) trackMap(tm *tMap) error {
	const op = "encrypt.(trackedMaps).trackMap"
	if tm == nil {
		return fmt.Errorf("%s: missing map: %w", op, ErrInvalidParameter)
	}
	if !tm.value.IsValid() {
		return fmt.Errorf("%s: map value is missing: %w", op, ErrInvalidParameter)
	}

	tmKind := tm.value.Kind()

	var isMapPtr bool
	if tmKind == reflect.Ptr && tm.value.Elem().Kind() == reflect.Map {
		isMapPtr = true
	}
	switch {
	case isMapPtr || tmKind == reflect.Map || tm.value.Type() == reflect.TypeOf(&structpb.Struct{}):
		maps[tm.value.Pointer()] = tm // we might need to move this if we ever change the value of tm to tm.Elem()
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
// haven't been tracked as filtered and it will mark them as filtered.  It will
// skip any fields within a map which have already been marked as filtered.
func (maps trackedMaps) processUnfiltered(ctx context.Context, ef *Filter, filterOverrides map[DataClassification]FilterOperation, opt ...Option) error {
	const op = "encrypt.(trackedMaps).processUnfiltered"
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

			var fPtr bool
			if field.Kind() == reflect.Ptr {
				field = field.Elem()
				fPtr = true
			}
			if field.Kind() == reflect.Interface {
				field = field.Elem()
			}

			ftype := field.Type()
			fkind := field.Kind()

			switch {
			// if the field is a string or []byte then we just need to sanitize it
			case ftype == reflect.TypeOf(""):
				s := field.String()
				f := reflect.Indirect(reflect.ValueOf(&s))
				if err := ef.filterValue(ctx, f, classificationTag, opt...); err != nil {
					return fmt.Errorf("%s: unable to filter string: %w", op, err)
				}
				v.SetMapIndex(key, f)

			case ftype == reflect.TypeOf([]uint8{}):
				s := field.Bytes()
				f := reflect.Indirect(reflect.ValueOf(&s))
				if err := ef.filterValue(ctx, f, classificationTag, opt...); err != nil {
					return fmt.Errorf("%s: unable to filter []byte: %w", op, err)
				}
				v.SetMapIndex(key, f)

			case ftype == reflect.TypeOf(wrapperspb.StringValue{}):
				s := field.FieldByName("Value").String()
				f := reflect.Indirect(reflect.ValueOf(&s))
				if err := ef.filterValue(ctx, f, classificationTag, opt...); err != nil {
					return fmt.Errorf("%s: unable to filter wrappers string value: %w", op, err)
				}
				vv := reflect.ValueOf(wrapperspb.StringValue{Value: s})
				v.SetMapIndex(key, vv)

			case ftype == reflect.TypeOf(wrapperspb.BytesValue{}):
				s := field.FieldByName("Value").Bytes()
				f := reflect.Indirect(reflect.ValueOf(&s))
				if err := ef.filterValue(ctx, f, classificationTag, opt...); err != nil {
					return fmt.Errorf("%s: unable to filter wrappers bytes value: %w", op, err)
				}
				vv := reflect.ValueOf(wrapperspb.BytesValue{Value: s})
				v.SetMapIndex(key, vv)

			case fkind == reflect.Slice:
				switch {
				// if the field is a slice of string or slice of []byte
				case ftype == reflect.TypeOf([]string{}) || ftype == reflect.TypeOf([][]uint8{}):
					if err := ef.filterSlice(ctx, classificationTag, field, opt...); err != nil {
						return fmt.Errorf("%s: unable to filter slice of strings: %w", op, err)
					}
				// if the field is a slice of structs, recurse through them...
				default:
					for i := 0; i < field.Len(); i++ {
						f := field.Index(i)
						if f.Kind() == reflect.Interface {
							f.Elem()
						}
						if f.Kind() == reflect.Ptr {
							f = f.Elem()
						}
						newMaps, err := newTrackedMaps()
						if err != nil {
							return fmt.Errorf("%s: unable to create new tracked maps for slice: %w", op, err)
						}
						switch {
						case f.Kind() == reflect.Struct:
							newOpts := make([]Option, 0, len(opt))
							newOpts = append(newOpts, opt...)
							newOpts = append(newOpts, withTrackedMaps(newMaps))
							if err := ef.filterField(ctx, f, filterOverrides, newOpts...); err != nil {
								return fmt.Errorf("%s: unable to filter slice of structs: %w", op, err)
							}

						case f.Kind() == reflect.Map:
							newMaps.trackMap(&tMap{
								value: f,
							})

						default:
							// for now, there's no default... perhaps in the
							// future there will be a reasonable default.
						}
						if err := newMaps.processUnfiltered(ctx, ef, filterOverrides, opt...); err != nil {
							return fmt.Errorf("%s: unable to process maps found in slice: %w", op, err)
						}
					}
				}

			case fkind == reflect.Struct:
				f := field
				if err := ef.filterField(ctx, f, filterOverrides, opt...); err != nil {
					return fmt.Errorf("%s: unable to filter struct: %w", op, err)
				}
				if fPtr {
					f = field.Addr()
				}
				v.SetMapIndex(key, f)

			case fkind == reflect.Map:
				newMaps, err := newTrackedMaps(&tMap{value: field})
				if err != nil {
					return fmt.Errorf("%s: unable to filter map: %w", op, err)
				}
				if err := newMaps.processUnfiltered(ctx, ef, filterOverrides, opt...); err != nil {
					return fmt.Errorf("%s: unable to process maps found in map: %w", op, err)
				}

			default:
				// at this point, there's no reasonable default.. wish there was.
			}

			// if you want to examine the "after filter" value you'll need to
			// look at the v.MapIndex(key) directly, not the field.
		}
		// very important to mark the current map as filtered before iterating
		m.filtered = true
		m.filteredFields = nil
	}
	return nil
}

func (maps trackedMaps) trackTaggable(taggable Taggable, pointer string) error {
	const (
		op            = "encrypt.(trackedMaps).trackTaggable"
		pathDelimiter = "/"
		badPath       = 1
		taggableMap   = 2
	)
	if taggable == nil {
		return fmt.Errorf("%s: missing taggable: %w", op, ErrInvalidParameter)
	}
	if pointer == "" {
		return fmt.Errorf("%s: missing pointer: %w", op, ErrInvalidParameter)
	}

	// need to determine what maps are referenced in this pointer tag, so spit
	// on the pointerstruct path delimiter of "/"
	segs := strings.Split(pointer, pathDelimiter)
	switch len(segs) {
	case badPath:
		return fmt.Errorf("%s: invalid taggable pointer: %w", op, ErrInvalidParameter)

	case taggableMap:
		// the path just pointed at a field within taggable
		ptr := reflect.ValueOf(taggable).Pointer()

		// Are we already tracking this map?
		if _, ok := maps[ptr]; !ok {
			v := reflect.ValueOf(taggable)
			// not sure if we need to worry if v.Kind() is a reflect.Ptr and
			// then get the elem... so for now, I'm going to skip that.
			tmap := &tMap{
				value:          v,
				filteredFields: map[string]struct{}{},
			}
			err := maps.trackMap(tmap)
			if err != nil {
				return fmt.Errorf("%s: unable to track taggable map: %w", op, err)
			}
		}
		// now, we're just going to mark a field referenced by the pointer
		// within the map as "filtered"
		maps[ptr].filteredFields[segs[len(segs)-1]] = struct{}{}

	default:
		// default is a map that we need to go get via the pointer
		foundMap, err := pointerstructure.Get(taggable, strings.Join(segs[:len(segs)-1], "/"))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		v := reflect.ValueOf(foundMap)
		ptr := v.Pointer()

		// Are we already tracking this map?
		if _, ok := maps[ptr]; !ok {
			tmap := &tMap{
				value:          v,
				filteredFields: map[string]struct{}{},
			}
			if err := maps.trackMap(tmap); err != nil {
				return fmt.Errorf("%s: unable to track map from pointer struct: %w", op, err)
			}
		}
		// now, we're just going to mark a field referenced by the pointer
		// within the map as "filtered"
		maps[ptr].filteredFields[segs[len(segs)-1]] = struct{}{}
	}
	return nil
}
