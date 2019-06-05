package kademlia

import (
	"fmt"
)

type NodeID uint64

type RoutingTable struct {
	root *bucket
}

func NewRoutingTable(k int, owner NodeID, max NodeID) *RoutingTable {
	return &RoutingTable{
		root: &bucket{
			K:     k,
			Low:   0,
			High:  max,
			Owner: owner,
			Nodes: []NodeID{},
		},
	}
}

// Insert returns true if the node is inserted and false otherwise.
func (t *RoutingTable) Insert(node NodeID) bool {
	if !t.root.WithinRange(node) {
		panic(fmt.Errorf("node id is outside of the entire id space"))
	}

	return t.root.Insert(node)
}

func (t *RoutingTable) Nodes() []NodeID {
	return t.root.AllNodes()
}

func (t *RoutingTable) Contains(node NodeID) bool {
	return t.root.Contains(node)
}

type bucket struct {
	K     int
	Low   NodeID
	High  NodeID
	Left  *bucket
	Right *bucket
	Owner NodeID
	Nodes []NodeID
}

func (b *bucket) AllNodes() []NodeID {
	if b.Left != nil && b.Right != nil {
		nodes := []NodeID{}
		nodes = append(nodes, b.Left.AllNodes()...)
		nodes = append(nodes, b.Right.AllNodes()...)
		return nodes
	}

	return b.Nodes
}

func (b *bucket) Contains(node NodeID) bool {
	if b.Left != nil && b.Right != nil {
		if b.Left.WithinRange(node) {
			return b.Left.Contains(node)
		}

		return b.Right.Contains(node)
	}

	for _, n := range b.Nodes {
		if node == n {
			return true
		}
	}

	return false
}

func (b *bucket) WithinRange(node NodeID) bool {
	return node >= b.Low && node <= b.High
}

// Insert returns true if the node is inserted and false otherwise.
func (b *bucket) Insert(node NodeID) bool {
	// If bucket has been split, recurse down the tree
	if b.Left != nil && b.Right != nil {
		if b.Left.WithinRange(node) {
			return b.Left.Insert(node)
		}

		return b.Right.Insert(node)
	}

	// If bucket is not full, insert is done
	if len(b.Nodes) < b.K {
		b.Nodes = append(b.Nodes, node)
		return true
	}

	// If owner of the bucket is not within range, ignore the node
	if !b.WithinRange(b.Owner) {
		return false
	}

	// Otherwise, split the bucket and reinsert all nodes
	b.split()

	for _, n := range b.Nodes {
		if b.Left.WithinRange(n) {
			b.Left.Insert(n)
		} else {
			b.Right.Insert(n)
		}
	}

	b.Nodes = []NodeID{}
	if b.Left.WithinRange(node) {
		return b.Left.Insert(node)
	}
	return b.Right.Insert(node)
}

func (b *bucket) split() {
	if b.Left != nil || b.Right != nil {
		panic(fmt.Errorf("bucket has already been split"))
	}

	mid := b.Low + ((b.High - b.Low) / 2)

	b.Left = &bucket{
		K:     b.K,
		Low:   b.Low,
		High:  mid,
		Owner: b.Owner,
		Nodes: []NodeID{},
	}
	b.Right = &bucket{
		K:     b.K,
		Low:   mid + 1,
		High:  b.High,
		Owner: b.Owner,
		Nodes: []NodeID{},
	}
}
