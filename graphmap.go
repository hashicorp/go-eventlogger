// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"fmt"
	"sync"
)

// TODO: remove this if Go ever introduces sync.Map with generics

// graphMap implements a type-safe synchronized map[PipelineID]*linkedNode
type graphMap struct {
	m sync.Map
}

// Range calls sync.Map.Range
func (g *graphMap) Range(f func(key PipelineID, value *linkedNode) bool) {
	g.m.Range(func(key, value interface{}) bool {
		return f(key.(PipelineID), value.(*linkedNode))
	})
}

// Store calls sync.Map.Store
func (g *graphMap) Store(id PipelineID, root *linkedNode) {
	g.m.Store(id, root)
}

// Delete calls sync.Map.Delete
func (g *graphMap) Delete(id PipelineID) {
	g.m.Delete(id)
}

// Nodes returns all the nodes referenced by the specified Pipeline
func (g *graphMap) Nodes(id PipelineID) ([]NodeID, error) {
	root, ok := g.m.Load(id)
	if !ok {
		return nil, fmt.Errorf("unable to load root node from underlying data store")
	}

	ln, ok := root.(*linkedNode)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve linked nodes from underlying data store")
	}
	nodes := ln.flatten()
	result := make([]NodeID, len(nodes))
	i := 0
	for k := range nodes {
		result[i] = k
		i++
	}
	return result, nil
}
