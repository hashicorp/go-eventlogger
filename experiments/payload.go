package experiments

// This package contains an experimental implementation of an efficient
// means of mutating the Payload of an Event.

import (
	"fmt"
	"strings"

	iradix "github.com/hashicorp/go-immutable-radix"
	//"github.com/mitchellh/pointerstructure"
)

// Path
type Path struct {
	Keys []string
}

func NewPath(keys ...string) *Path {
	return &Path{
		Keys: keys,
	}
}

func (p *Path) String() string {
	var str strings.Builder
	for i, k := range p.Keys {
		if i > 0 {
			str.WriteString("/")
		}
		str.WriteString(k)
	}
	return str.String()
}

// Payload is the data payload in an Event. Payload is immutable, and can
// create new versions of itself efficiently.  A Payload cannot have existing
// keys deleted, nor can it have new keys added to it.
type Payload interface {
	Get(path *Path) (interface{}, bool)
	Set(path *Path, val interface{}) (Payload, error)
}

type payload struct {
	base map[string]interface{}
	tree *iradix.Tree
}

// NewPayload creates a new Payload. Once a Payload is created, the underlying map
// that is passed in must never be mutated.
func NewPayload(m map[string]interface{}) Payload {
	return &payload{
		base: m,
		tree: iradix.New(),
	}
}

func (p *payload) Get(path *Path) (interface{}, bool) {

	val, ok := p.tree.Get([]byte(path.String()))
	if ok {
		return val, true
	}

	return p.getBase(path)
}

func (p *payload) getBase(path *Path) (interface{}, bool) {

	last := len(path.Keys) - 1
	curMap := p.base
	for i := 0; i < last; i++ {
		val, ok := curMap[path.Keys[i]]
		if !ok {
			return nil, false
		}
		curMap, ok = val.(map[string]interface{})
		if !ok {
			return nil, false
		}
	}

	val, ok := curMap[path.Keys[last]]
	return val, ok
}

func (p *payload) Set(path *Path, val interface{}) (Payload, error) {

	if _, ok := p.getBase(path); !ok {
		return nil, fmt.Errorf("Cannot set new field %s", path.String())
	}

	tree, _, _ := p.tree.Insert([]byte(path.String()), val)
	return &payload{
		base: p.base,
		tree: tree,
	}, nil
}
