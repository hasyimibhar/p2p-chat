package kademlia

import (
	"fmt"
	"math/rand"
	"sort"
)

type NodeID uint64

type RoutingEntry struct {
	NodeID   NodeID
	NodeAddr string
}

// RoutingTable is a Kademlia routing table.
type RoutingTable struct {
	root *bucket
}

// NewRoutingTable creates a Kademlia routing table.
func NewRoutingTable(k int, owner NodeID, max NodeID) *RoutingTable {
	return &RoutingTable{
		root: &bucket{
			K:       k,
			Low:     0,
			High:    max,
			Owner:   owner,
			Entries: []RoutingEntry{},
		},
	}
}

// Insert returns true if the node is inserted and false otherwise.
func (t *RoutingTable) Insert(entry RoutingEntry) bool {
	if !t.root.WithinRange(entry.NodeID) {
		panic(fmt.Errorf("node id is outside of the entire id space"))
	}

	return t.root.Insert(entry)
}

// Nodes returns list of nodes inside the routing table.
func (t *RoutingTable) Nodes() []RoutingEntry {
	return t.root.AllNodes()
}

// DistantNodes returns at most n nodes from a distant subtree randomly.
func (t *RoutingTable) DistantNodes(n int) []RoutingEntry {
	var nodes []RoutingEntry

	if t.root.Left != nil && t.root.Right != nil {
		if t.root.Left.Contains(t.root.Owner) {
			nodes = t.root.Right.AllNodes()
		} else {
			nodes = t.root.Left.AllNodes()
		}
	} else {
		nodes = t.root.Entries
	}

	if len(nodes) < n {
		return nodes
	}

	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	return nodes[:n]
}

// FindNode returns at most k nodes that is closest to the target node.
func (t *RoutingTable) FindNode(target NodeID) []RoutingEntry {
	return t.findClosest(target, t.root.K)
}

// FindClosest returns at most n nodes that is closest to the owner node.
func (t *RoutingTable) FindClosest(n int) []RoutingEntry {
	return t.findClosest(t.root.Owner, n)
}

func (t *RoutingTable) findClosest(target NodeID, n int) []RoutingEntry {
	candidates := t.root.AllNodes()

	sort.Sort(ByXORDistance{candidates, target})

	if len(candidates) < n {
		return candidates
	}
	return candidates[:n]
}

// Contains returns true if the node id exists inside the
// routing table.
func (t *RoutingTable) Contains(node NodeID) bool {
	return t.root.Contains(node)
}

// bucket is a Kademlia k-bucket.
type bucket struct {
	K       int
	Low     NodeID
	High    NodeID
	Left    *bucket
	Right   *bucket
	Owner   NodeID
	Entries []RoutingEntry
}

// AllNodes returns the list of all nodes inside the bucket
// and its child buckets.
func (b *bucket) AllNodes() []RoutingEntry {
	if b.Left != nil && b.Right != nil {
		nodes := []RoutingEntry{}
		nodes = append(nodes, b.Left.AllNodes()...)
		nodes = append(nodes, b.Right.AllNodes()...)
		return nodes
	}

	return b.Entries
}

// Contains returns true if the node id exists inside the bucket
// or its child buckets.
func (b *bucket) Contains(node NodeID) bool {
	if b.Left != nil && b.Right != nil {
		if b.Left.WithinRange(node) {
			return b.Left.Contains(node)
		}

		return b.Right.Contains(node)
	}

	for _, n := range b.Entries {
		if node == n.NodeID {
			return true
		}
	}

	return false
}

// WithinRange returns wether the node id falls within the bucket range.
func (b *bucket) WithinRange(node NodeID) bool {
	return node >= b.Low && node <= b.High
}

// Insert returns true if the node is inserted and false otherwise.
func (b *bucket) Insert(entry RoutingEntry) bool {
	// If bucket has been split, recurse down the tree
	if b.Left != nil && b.Right != nil {
		if b.Left.WithinRange(entry.NodeID) {
			return b.Left.Insert(entry)
		}

		return b.Right.Insert(entry)
	}

	// If bucket is not full, insert is done
	if len(b.Entries) < b.K {
		b.Entries = append(b.Entries, entry)
		return true
	}

	// If owner of the bucket is not within range, ignore the node
	if !b.WithinRange(b.Owner) {
		return false
	}

	// Otherwise, split the bucket and reinsert all nodes
	b.split()

	for _, n := range b.Entries {
		if b.Left.WithinRange(n.NodeID) {
			b.Left.Insert(n)
		} else {
			b.Right.Insert(n)
		}
	}

	b.Entries = []RoutingEntry{}
	if b.Left.WithinRange(entry.NodeID) {
		return b.Left.Insert(entry)
	}
	return b.Right.Insert(entry)
}

// split splits the bucket into 2 halves.
func (b *bucket) split() {
	if b.Left != nil || b.Right != nil {
		panic(fmt.Errorf("bucket has already been split"))
	}

	mid := b.Low + ((b.High - b.Low) / 2)

	b.Left = &bucket{
		K:       b.K,
		Low:     b.Low,
		High:    mid,
		Owner:   b.Owner,
		Entries: []RoutingEntry{},
	}
	b.Right = &bucket{
		K:       b.K,
		Low:     mid + 1,
		High:    b.High,
		Owner:   b.Owner,
		Entries: []RoutingEntry{},
	}
}

type RoutingEntries []RoutingEntry

func (e RoutingEntries) Len() int      { return len(e) }
func (e RoutingEntries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// ByXORDistance sorts the list of routing entries by increasing XOR distance
// from the target node.
type ByXORDistance struct {
	RoutingEntries
	Target NodeID
}

func (e ByXORDistance) Less(i, j int) bool {
	return e.RoutingEntries[i].NodeID^e.Target < e.RoutingEntries[j].NodeID^e.Target
}
