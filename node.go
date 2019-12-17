package eventlogger

type Node interface {
	Process(g Graph, e Event) error
}

//type LinkedNode struct {
//	Next Node
//}

//// LinkedNode creates a linked list, a.k.a “Pipeline”.
//// Redactors, Filters and Writers are likely to use this struct.
//struct LinkedNode {
//	next Node
//}
//
//// FanOutNode creates a tree, which will be
//// useful for e.g. fanning out to multiple Sinks.
//struct FanOutNode {
//	next []Nodes
//}
//
//// A Leaf in the graph.  Sinks should be formally defined
//// as being a Leaf.
//struct Leaf {
//}
