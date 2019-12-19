package imap

import (
	iradix "github.com/hashicorp/go-immutable-radix"
)

// IMap is an immutable data structure that layers modifications onto a
// base map[string]interface{}.
type IMap struct {
	base    map[string]interface{}
	added   *iradix.Tree
	deleted *iradix.Tree
}

// Txn is a transaction on an IMap.
type Txn struct {
	base    map[string]interface{}
	added   *iradix.Tree
	deleted *iradix.Tree
}

// NewIMap creates a new IMap. Once an IMap is created, the underlying base map
// that is passed in must never be mutated.
func NewIMap(base map[string]interface{}) *IMap {
	return &IMap{
		base:    base,
		added:   iradix.New(),
		deleted: iradix.New(),
	}
}

// Get looks up the value for a specified path, along with whether
// the value was found.
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

	return getBase(m.base, path)
}

func getBase(base map[string]interface{}, path *Path) (interface{}, bool) {

	last := len(path.Keys) - 1
	cur := base
	for i := 0; i < last; i++ {
		val, ok := cur[path.Keys[i]]
		if !ok {
			return nil, false
		}
		cur, ok = val.(map[string]interface{})
		if !ok {
			return nil, false
		}
	}

	val, ok := cur[path.Keys[last]]
	return val, ok
}

// Set creates a new IMap that has a new value at the specified Path.
func (m *IMap) Set(path *Path, val interface{}) *IMap {
	tx := m.Txn()
	tx.Set(path, val)
	return tx.Commit()
}

// Delete creates a new IMap that no longer has a value at the specified Path.
func (m *IMap) Delete(path *Path) *IMap {
	tx := m.Txn()
	tx.Delete(path)
	return tx.Commit()
}

// Txn creates a new Transaction.
func (m *IMap) Txn() *Txn {
	return &Txn{
		base:    m.base,
		added:   m.added,
		deleted: m.deleted,
	}
}

// Get looks up the value for a specified path, along with whether
// the value was found.
func (t *Txn) Get(path *Path) (interface{}, bool) {

	k := []byte(path.String())

	_, ok := t.deleted.Get(k)
	if ok {
		return nil, false
	}
	val, ok := t.added.Get(k)
	if ok {
		return val, true
	}

	return getBase(t.base, path)
}

// Set a value at the specified path in the transaction.
func (t *Txn) Set(path *Path, val interface{}) {
	k := []byte(path.String())
	a, _, _ := t.added.Insert(k, val)
	d, _, _ := t.deleted.Delete(k)
	t.added = a
	t.deleted = d
}

// Delete the path from the transaction
func (t *Txn) Delete(path *Path) {
	k := []byte(path.String())
	a, _, _ := t.added.Delete(k)
	d, _, _ := t.deleted.Insert(k, true)
	t.added = a
	t.deleted = d
}

// Commit commits the transaction, returning a new IMap
func (t *Txn) Commit() *IMap {
	return &IMap{
		base:    t.base,
		added:   t.added,
		deleted: t.deleted,
	}
}
