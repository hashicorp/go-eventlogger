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
	// Store the root node and increment how many we have.
	// NOTE: These two actions might not be atomic, so potentially something could
	// start to range over the map before we've made the change to the total number
	// of roots.
	g.m.Store(id, root)
	g.numRoots++
}

// Delete calls sync.Map.Delete
func (g *graphMap) Delete(id PipelineID) {
	// Delete the nodes for the pipeline, and decrement how many root nodes we have.
	// NOTE: These two actions might not be atomic, so potentially something could
	// start to range over the map before we've made the change to the total number
	// of roots.
	g.m.Delete(id)
	g.numRoots--
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
