package node

import "testing"

type fakeFSNode struct {
	n Node
}

func (f *fakeFSNode) GetNode() *Node {
	return &f.n
}

func TestFSNodeInterface(t *testing.T) {
	f := &fakeFSNode{n: Node{ID: 1, Name: "test"}}

	var fs FSNode = f
	if fs.GetNode().Name != "test" {
		t.Fatalf("expected %q, got %q", "test", fs.GetNode().Name)
	}
}