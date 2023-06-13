package eventlogger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodes_ListNodes_UnregisteredPipeline(t *testing.T) {
	g := &graphMap{}
	ids, err := g.Nodes(PipelineID("31"))
	require.Error(t, err)
	require.EqualError(t, err, "unable to load root node from underlying data store")
	require.Nil(t, ids)
}

func TestNodes_ListNodes_RegisteredPipeline(t *testing.T) {
	g := &graphMap{}

	// Create some nodes
	ids := []NodeID{"a", "b", "c"}
	nodes := []Node{
		&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
	}

	linkedNodes, err := linkNodes(nodes, ids)
	require.NoError(t, err)

	g.Store(PipelineID("1"), linkedNodes)
	nodeIDs, err := g.Nodes(PipelineID("1"))
	require.NoError(t, err)
	require.NotNil(t, nodeIDs)
	require.Contains(t, nodeIDs, NodeID("a"))
	require.Contains(t, nodeIDs, NodeID("b"))
	require.Contains(t, nodeIDs, NodeID("c"))
}
