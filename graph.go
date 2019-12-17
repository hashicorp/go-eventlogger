package eventlogger

import "errors"

// Graph
type Graph struct {
	Root Node
}

//// Process processes the Event by routing it through all of the graph's nodes,
//// starting with the root node.
//func (g *Graph) Process(e *Event) error {
//
//	node := g.Root
//
//	for node != nil {
//
//		// Process the current Node
//		err := node.Process(e)
//		if err != nil {
//			return err
//		}
//
//		// Go to the next Node
//		if ln, ok := node.(Linkable); ok {
//			node = ln.Next()
//		}
//	}
//
//	return nil
//}

// LinkNodes is a convenience function that connects
// Nodes together into a linked list. All of the nodes except the
// last one must be LinkableNodes
func LinkNodes(nodes []Node) ([]Node, error) {

	num := len(nodes)
	if num < 2 {
		return nodes, nil
	}

	for i := 0; i < num-1; i++ {
		ln, ok := nodes[i].(LinkableNode)
		if !ok {
			return nil, errors.New("Node is not Linkable")
		}
		ln.SetNext([]Node{nodes[i+1]})
	}

	return nodes, nil
}
