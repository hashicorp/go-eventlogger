package imap

import (
	iradix "github.com/hashicorp/go-immutable-radix"
)

// IMap is an immutable data structure that layers changes onto a
// base map[string]interface{}.
type IMap struct {
	base    map[string]interface{}
	added   *iradix.Tree
	deleted *iradix.Tree
}

//// Txn is a transaction on an IMap.
//type Txn interface {
//	Get(path *Path) (interface{}, bool)
//	Set(path *Path, val interface{})
//	Delete(path *Path)
//}

// NewIMap creates a new IMap. Once an IMap is created, the underlying
// map that is passed in must never be mutated.
func NewIMap(base map[string]interface{}) *IMap {
	return &IMap{
		base:    base,
		added:   iradix.New(),
		deleted: iradix.New(),
	}
}

func (m *IMap) Get(path *Path) (interface{}, bool) {

	k := []byte(path.String())

	_, ok := m.deleted.Get(k)
	if ok {
		return nil, false
	}

	val, ok := m.added.Get(k)
	if ok {
		return val, true
	}

	return m.getBase(path)
}

func (m *IMap) getBase(path *Path) (interface{}, bool) {

	last := len(path.Keys) - 1
	curMap := m.base
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

func (m *IMap) Set(path *Path, val interface{}) *IMap {

	k := []byte(path.String())

	added, _, _ := m.added.Insert(k, val)
	deleted, _, _ := m.deleted.Delete(k)

	return &IMap{
		base:    m.base,
		added:   added,
		deleted: deleted,
	}
}

func (m *IMap) Delete(path *Path) *IMap {

	k := []byte(path.String())

	added, _, _ := m.added.Delete(k)
	deleted, _, _ := m.deleted.Insert(k, true)

	return &IMap{
		base:    m.base,
		added:   added,
		deleted: deleted,
	}
}
