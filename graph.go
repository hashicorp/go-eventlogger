package eventlogger

// Graph
type Graph struct {
	Root Node
}

// Process the Envelope by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(e *Envelope) error {
	return process(g.Root)
}

// Recursively process every node in the graph.
func (g *Graph) process(node Node, env *Envelope) error {

	// Process the current Node
	env, err := node.Process(env)
	if err != nil {
		return err
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(Linkable); ok {
		children := ln.Next()
		for i := 0; i < len(children); i++ {
			process(children[i], env)
		}
	}
}
