package eventlogger

import "errors"

// Graph
type Graph struct {
	Root Node
}

// Process processes the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(e *Event) error {

	node := g.Root

	for node != nil {

		// Process the current Node
		err := node.Process(e)
		if err != nil {
			return err
		}

		// Go to the next Node
		if ln, ok := node.(Linkable); ok {
			node = ln.Next()
		}
	}

	return nil
}

// LinkNodes is a convenience function that links
// together Nodes into a linked list.  All of the
// Nodes except the last one must be Linkable.
func LinkNodes(nodes []Node) error {

	num := len(nodes)
	if num < 2 {
		return nil
	}

	for i := 0; i < num-1; i++ {
		ln, ok := nodes[i].(Linkable)
		if !ok {
			return errors.New("Node is not Linkable")
		}
		ln.SetNext(nodes[i+1])
	}

	return nil
}
