package eventlogger

// Graph
type Graph struct {
	Root Node
}

// Process processes the Envelope by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(e *Envelope) error {
	//process(g.Root)
	return nil
}

//func (g *Graph) process(node Node, e *Envelope) error {
//
//	//	// Process the current Node
//	//	err := node.Process(e)
//	//	if err != nil {
//	//		return err
//	//	}
//	//
//	//	// Process any child nodes
//	//	if ln, ok := node.(Linkable); ok {
//	//		children := ln.Next()
//	//		for i := 0; i < len(children); i++ {
//	//			process(children[i], e)
//	//		}
//	//	}
//}
