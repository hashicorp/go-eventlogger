// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestNodes_ListNodes_UnregisteredPipeline checks that we get the right error
// when attempting to get the linked nodes for an unregistered pipeline.
func TestNodes_ListNodes_UnregisteredPipeline(t *testing.T) {
	g := &graphMap{}
	ids, err := g.Nodes(PipelineID("31"))
	require.Error(t, err)
	require.EqualError(t, err, "unable to load root node from underlying data store")
	require.Nil(t, ids)
}

// TestNodes_ListNodes_RegisteredPipeline checks that we can retrieve all registered
// nodes referenced by a registered pipeline.
func TestNodes_ListNodes_RegisteredPipeline(t *testing.T) {
	g := &graphMap{}

	// Create some nodes
	ids := []NodeID{"a", "b", "c"}
	nodes := []Node{
		&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
	}

	linkedNodes, err := linkNodes(nodes, ids)
	require.NoError(t, err)

	reg := &registeredPipeline{rootNode: linkedNodes, registrationPolicy: AllowOverwrite}
	g.Store(PipelineID("1"), reg)
	nodeIDs, err := g.Nodes(PipelineID("1"))
	require.NoError(t, err)
	require.NotNil(t, nodeIDs)
	require.Contains(t, nodeIDs, NodeID("a"))
	require.Contains(t, nodeIDs, NodeID("b"))
	require.Contains(t, nodeIDs, NodeID("c"))
}

func TestGraphMap_Store(t *testing.T) {
	t.Parallel()

	g := &graphMap{}
	findPipeline := pipelineFinder(g)

	// Set up pipeline for storing.
	id := PipelineID("foo")
	p := &registeredPipeline{
		registrationPolicy: "bar",
	}

	// Sanity check we have nothing to start with.
	require.Equal(t, 0, g.numRoots)
	v := findPipeline(id)
	require.Nil(t, v)

	// Store the pipeline then check we stored it and incremented the counter.
	g.Store(id, p)
	require.Equal(t, 1, g.numRoots)
	v = findPipeline(id)
	require.NotNil(t, v)
	require.Equal(t, RegistrationPolicy("bar"), v.registrationPolicy)

	// Store it again, and check it's still there but the counter hasn't changed
	// since it's per distinct ID.
	p.registrationPolicy = "baz"
	g.Store(id, p)
	require.Equal(t, 1, g.numRoots)
	v = findPipeline(id)
	require.NotNil(t, v)
	require.Equal(t, RegistrationPolicy("baz"), v.registrationPolicy)
}

func TestGraphMap_Delete(t *testing.T) {
	t.Parallel()

	g := &graphMap{}
	findPipeline := pipelineFinder(g)

	// Set up pipeline for storing.
	id := PipelineID("foo")
	p := &registeredPipeline{
		registrationPolicy: "bar",
	}

	// Sanity check we have nothing to start with.
	require.Equal(t, 0, g.numRoots)
	v := findPipeline(id)
	require.Nil(t, v)

	// Store the pipeline then check we stored it and incremented the counter.
	g.Store(id, p)
	require.Equal(t, 1, g.numRoots)
	v = findPipeline(id)
	require.NotNil(t, v)
	require.Equal(t, RegistrationPolicy("bar"), v.registrationPolicy)

	// Now delete the pipeline and check it's gone and the counter went down.
	g.Delete(id)
	require.Equal(t, 0, g.numRoots)
	v = findPipeline(id)
	require.Nil(t, v)

	// Delete again and make sure nothing funky happens.
	g.Delete(id)
	require.Equal(t, 0, g.numRoots)
	v = findPipeline(id)
	require.Nil(t, v)
}

// pipelineFinder returns a func that can be used to find a pipeline in a graphMap.
func pipelineFinder(g *graphMap) func(id PipelineID) *registeredPipeline {
	return func(id PipelineID) *registeredPipeline {
		var res *registeredPipeline

		g.Range(func(key PipelineID, rp *registeredPipeline) bool {
			if key == id {
				res = rp
				return false
			}
			return true
		})

		return res
	}
}
