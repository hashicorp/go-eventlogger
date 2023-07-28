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

	reg := &pipelineRegistration{rootNode: linkedNodes, registrationPolicy: AllowOverwrite}
	g.Store(PipelineID("1"), reg)
	nodeIDs, err := g.Nodes(PipelineID("1"))
	require.NoError(t, err)
	require.NotNil(t, nodeIDs)
	require.Contains(t, nodeIDs, NodeID("a"))
	require.Contains(t, nodeIDs, NodeID("b"))
	require.Contains(t, nodeIDs, NodeID("c"))
}
