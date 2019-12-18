package eventdata

//import (
//	"fmt"
//	"strings"
//)
//
//// This is experimental!!! And its not even for MVP anyway!
//
////type Visitor func(path []string, value interface{})
////
////type JSONPointer
////
////type EventData interface {
////	Visit(Visitor)
////	Get(jsonPtr string) interface{}
////	Set(jsonPtr string, value string)
////}
////
//
//func NewFromMap(input map[string]interface{}) EventPayload {
//	//...
//}
//
//// Parse parses a pointer from the input string. The input string
//// is expected to follow the format specified by RFC 6901: '/'-separated
//// parts. Each part can contain escape codes to contain '/' or '~'.
//func Parse(input string) (*Pointer, error) {
//	// Special case the empty case
//	if input == "" {
//		return &Pointer{}, nil
//	}
//
//	// We expect the first character to be "/"
//	if input[0] != '/' {
//		return nil, fmt.Errorf(
//			"parse Go pointer %q: first char must be '/'", input)
//	}
//
//	// Trim out the first slash so we don't have to +1 every index
//	input = input[1:]
//
//	// Parse out all the parts
//	var parts []string
//	lastSlash := -1
//	for i, r := range input {
//		if r == '/' {
//			parts = append(parts, input[lastSlash+1:i])
//			lastSlash = i
//		}
//	}
//
//	// Add last part
//	parts = append(parts, input[lastSlash+1:])
//
//	// Process each part for string replacement
//	for i, p := range parts {
//		// Replace ~1 followed by ~0 as specified by the RFC
//		parts[i] = strings.Replace(
//			strings.Replace(p, "~1", "/", -1), "~0", "~", -1)
//	}
//
//	return &Pointer{Parts: parts}, nil
//}
//
//// MustParse is like Parse but panics if the input cannot be parsed.
//func MustParse(input string) *Pointer {
//	p, err := Parse(input)
//	if err != nil {
//		panic(err)
//	}
//
//	return p
//}
//
//// Pointer represents a pointer to a specific value. You can construct
//// a pointer manually or use Parse.
//type Pointer struct {
//	// Parts are the pointer parts. No escape codes are processed here.
//	// The values are expected to be exact. If you have escape codes, use
//	// the Parse functions.
//	Parts []string
//}
//
////func (p *Pointer) Delete(s interface{}) (interface{}, error)
////func (p *Pointer) Get(v interface{}) (interface{}, error)
////func (p *Pointer) IsRoot() bool
////func (p *Pointer) Parent() *Pointer
////func (p *Pointer) Set(s, v interface{}) (interface{}, error)
////func (p *Pointer) String() string
