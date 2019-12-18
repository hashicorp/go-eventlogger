package eventlogger

// Graph
type Graph struct {
	Root Node
}

// Process the Event by routing it through all of the graph's nodes,
// starting with the root node.
func (g *Graph) Process(env *Event) error {
	return g.process(g.Root, env)
}

// Recursively process every node in the graph.
func (g *Graph) process(node Node, env *Event) error {

	// Process the current Node
	env, err := node.Process(env)
	if err != nil {
		return err
	}

	// If the new Event is nil, it has been filtered out and we are done.
	if env == nil {
		return nil
	}

	// Process any child nodes.  This is depth-first.
	if ln, ok := node.(LinkableNode); ok {
		for _, child := range ln.Next() {

			err = g.process(child, env)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
