// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"
)

// TestLinkNodes ensures that we are able to create a graph of linked nodes correctly.
// NOTE: This test should not be run in parallel as it sets a package level variable
// on 'deep' to ensure we compare unexported fields too.
func TestLinkNodes(t *testing.T) {
	n1, n2, n3 := &Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"}
	root, err := linkNodes([]Node{n1, n2, n3}, []NodeID{"1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}

	expected := &linkedNode{
		node:   n1,
		nodeID: "1",
		next: []*linkedNode{{
			node:   n2,
			nodeID: "2",
			next: []*linkedNode{{
				node:   n3,
				nodeID: "3",
			}},
		}},
	}

	deep.CompareUnexportedFields = true
	t.Cleanup(func() { deep.CompareUnexportedFields = false })

	if diff := deep.Equal(root, expected); len(diff) > 0 {
		t.Fatal(diff)
	}
}

// TestLinkNodesErrors attempts to exercise the linkNodes func such that we hit
// the early return error checking on the incoming parameters.
func TestLinkNodesErrors(t *testing.T) {
	tests := map[string]struct {
		nodes            []Node
		ids              []NodeID
		wantErrorMessage string
	}{
		"nil-nodes": {
			nodes:            nil,
			ids:              []NodeID{"1", "2", "3"},
			wantErrorMessage: "no nodes given",
		},
		"no-nodes": {
			nodes:            []Node{},
			ids:              []NodeID{"1", "2", "3"},
			wantErrorMessage: "no nodes given",
		},
		"nil-ids": {
			nodes: []Node{
				&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
			},
			ids:              nil,
			wantErrorMessage: "no IDs given",
		},
		"no-ids": {
			nodes: []Node{
				&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
			},
			ids:              []NodeID{},
			wantErrorMessage: "no IDs given",
		},
		"more-nodes-than-ids": {
			nodes: []Node{
				&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
			},
			ids:              []NodeID{"1", "2"},
			wantErrorMessage: "number of nodes does not match number of IDs",
		},
		"less-nodes-than-ids": {
			nodes: []Node{
				&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
			},
			ids:              []NodeID{"1", "2", "3", "4"},
			wantErrorMessage: "number of nodes does not match number of IDs",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := linkNodes(tc.nodes, tc.ids)
			require.Error(t, err)
			require.EqualError(t, err, tc.wantErrorMessage)
		})
	}
}

// TestFlattenNodes tests that given a 'root' node we can correctly flatten it
// out to retrieve the NodeIDs of linked nodes.
func TestFlattenNodes_LinkNodes(t *testing.T) {
	ids := []NodeID{"1", "2", "3"}
	nodes := []Node{
		&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
	}

	linkedNodes, err := linkNodes(nodes, ids)
	require.NoError(t, err)

	flatNodes := linkedNodes.flatten()
	require.Contains(t, flatNodes, NodeID("1"))
	require.Contains(t, flatNodes, NodeID("2"))
	require.Contains(t, flatNodes, NodeID("3"))
	require.Equal(t, 3, len(flatNodes))
}

// TestFlattenNodes_LinkNodesAndSinks tests that given a more complex set of linked
// nodes we can still get the right set of registered nodes.
func TestFlattenNodes_LinkNodesAndSinks(t *testing.T) {
	ids := []NodeID{"1", "2"}
	nodes := []Node{
		&Filter{Predicate: nil}, &JSONFormatter{},
	}

	sinkIds := []NodeID{"x", "y", "z"}
	sinkNodes := []Node{
		&FileSink{Path: "test.log"}, &FileSink{Path: "foo.log"}, &FileSink{Path: "bar.log"},
	}

	linkedNodes, err := linkNodesAndSinks(nodes, sinkNodes, ids, sinkIds)
	require.NoError(t, err)

	flatNodes := linkedNodes.flatten()
	require.Contains(t, flatNodes, NodeID("1"))
	require.Contains(t, flatNodes, NodeID("2"))
	require.Contains(t, flatNodes, NodeID("x"))
	require.Contains(t, flatNodes, NodeID("y"))
	require.Contains(t, flatNodes, NodeID("z"))
	require.Equal(t, 5, len(flatNodes))
}
