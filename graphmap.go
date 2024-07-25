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

	// numRoots attempts to track the number of root nodes that are registered in
	// the associate map of pipelines. This can be useful for the graph to decide
	// how large a channel should be for receiving Status from nodes as they process.
	// Later it may require a lock/mutex in order to synchronize the Store and Delete
	// operations on the map, but for now this should be accurate enough.
	numRoots int
}

// registeredPipeline represents both linked nodes and the registration policy
// for the pipeline.
type registeredPipeline struct {
	rootNode           *linkedNode
	registrationPolicy RegistrationPolicy
}

// Range calls sync.Map.Range
func (g *graphMap) Range(f func(key PipelineID, value *registeredPipeline) bool) {
	g.m.Range(func(key, value interface{}) bool {
		return f(key.(PipelineID), value.(*registeredPipeline))
	})
}

// Store calls sync.Map.Store
func (g *graphMap) Store(id PipelineID, root *registeredPipeline) {
	// Store the root node and increment how many we have (if this is a new pipeline).
	// NOTE: These two actions might not be atomic, so potentially something could
	// start to range over the map before we've made the change to the total number
	// of roots.
	if !g.Exists(id) {
		g.numRoots++
	}
	g.m.Store(id, root)
}

// Delete calls sync.Map.Delete
func (g *graphMap) Delete(id PipelineID) {
	if !g.Exists(id) {
		return
	}

	// Delete the root node for the pipeline if it was already stored, and decrement
	// how many we have.
	// NOTE: These two actions might not be atomic, so potentially something could
	// start to range over the map before we've made the change to the total number
	// of roots.
	g.numRoots--
	g.m.Delete(id)
}

// Nodes returns all the nodes referenced by the specified Pipeline
func (g *graphMap) Nodes(id PipelineID) ([]NodeID, error) {
	v, ok := g.m.Load(id)
	if !ok {
		return nil, fmt.Errorf("unable to load root node from underlying data store")
	}

	pr, ok := v.(*registeredPipeline)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve pipeline registration (linked nodes and policy) from underlying data store")
	}

	nodes := pr.rootNode.flatten()
	result := make([]NodeID, len(nodes))
	i := 0
	for k := range nodes {
		result[i] = k
		i++
	}

	return result, nil
}

// Exists determines whether a PipelineID is already stored within the graphMap.
func (g *graphMap) Exists(id PipelineID) bool {
	var found bool

	g.Range(func(key PipelineID, v *registeredPipeline) bool {
		if key == id {
			found = true
			return false
		}
		return true
	})

	return found
}
