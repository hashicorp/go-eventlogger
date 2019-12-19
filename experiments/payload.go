package experiments

// This package contains an experimental implementation of an efficient
// means of mutating the Payload of an Event.

import (
	"strings"

	iradix "github.com/hashicorp/go-immutable-radix"
	//"github.com/mitchellh/pointerstructure"
)

// Path is a path in a Payload.
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
	for i, key := range p.Keys {
		if i > 0 {
			str.WriteString("/")
		}
		str.WriteString(key)
	}
	return str.String()
}

// Payload is the data payload in an Event. Payload is immutable, and can
// create new versions of itself efficiently.
type Payload interface {
	Get(path *Path) (interface{}, bool)
	Set(path *Path, val interface{}) Payload
	Delete(path *Path) Payload
}

type payload struct {
	base    map[string]interface{}
	tree    *iradix.Tree
	deleted *iradix.Tree
}

// NewPayload creates a new Payload. Once a Payload is created, the underlying
// map that is passed in must never be mutated.
func NewPayload(m map[string]interface{}) Payload {
	return &payload{
		base:    m,
		tree:    iradix.New(),
		deleted: iradix.New(),
	}
}

func (p *payload) Get(path *Path) (interface{}, bool) {

	k := []byte(path.String())

	_, ok := p.deleted.Get(k)
	if ok {
		return nil, false
	}

	val, ok := p.tree.Get(k)
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

func (p *payload) Set(path *Path, val interface{}) Payload {

	k := []byte(path.String())

	tree, _, _ := p.tree.Insert(k, val)
	deleted, _, _ := p.deleted.Delete(k)

	return &payload{
		base:    p.base,
		tree:    tree,
		deleted: deleted,
	}
}

func (p *payload) Delete(path *Path) Payload {

	k := []byte(path.String())

	tree, _, _ := p.tree.Delete(k)
	deleted, _, _ := p.deleted.Insert(k, true)

	return &payload{
		base:    p.base,
		tree:    tree,
		deleted: deleted,
	}
}
