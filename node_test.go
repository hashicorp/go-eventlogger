// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventlogger

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/go-test/deep"
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
func TestFlattenNodes(t *testing.T) {
	ids := []NodeID{"1", "2", "3"}
	nodes := []Node{
		&Filter{Predicate: nil}, &JSONFormatter{}, &FileSink{Path: "test.log"},
	}

	linkedNodes, err := linkNodes(nodes, ids)
	require.NoError(t, err)

	flatNodes := linkedNodes.flatten(nil)
	require.Contains(t, flatNodes, NodeID("1"))

}
