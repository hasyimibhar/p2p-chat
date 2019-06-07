package kademlia

import (
	"math/rand"
	"testing"
	"time"
)

func TestRoutingTable_Insert(t *testing.T) {
	table := NewRoutingTable(1, 13, 15)

	nodes := []struct {
		ID       NodeID
		Inserted bool
	}{
		{1, true},
		{10, true},
		{6, false},
		{9, false},
		{12, true},
		{15, true},
	}

	for _, n := range nodes {
		inserted := table.Insert(RoutingEntry{n.ID, ""})
		if inserted != n.Inserted {
			t.Fatal("node inserted is wrong")
		}
	}

	for _, n := range nodes {
		if table.Contains(n.ID) != n.Inserted {
			t.Fatal("node list is wrong")
		}
	}

	allNodes := table.Nodes()

	for _, n := range nodes {
		if n.Inserted {
			found := false
			for _, nn := range allNodes {
				if nn.NodeID == n.ID {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("node list is wrong, doesn't contain %d", n.ID)
			}
		} else {
			for _, nn := range allNodes {
				if nn.NodeID == n.ID {
					t.Fatalf("node list is wrong, contains %d", n.ID)
				}
			}
		}
	}
}

func TestRoutingTable_DistantNodes(t *testing.T) {
	table := NewRoutingTable(1, 13, 15)

	nodes := []struct {
		ID NodeID
	}{
		{1},
		{10},
		{6},
		{9},
		{12},
		{15},
	}

	for _, n := range nodes {
		table.Insert(RoutingEntry{n.ID, ""})
	}

	distant := table.DistantNodes(3)
	if len(distant) != 1 {
		t.Fatal("there should only be 1 distant node")
	}
	if distant[0].NodeID != NodeID(1) {
		t.Fatal("wrong distant node")
	}
}

func TestRoutingTable_FindNode(t *testing.T) {
	table := NewRoutingTable(1, 13, 15)

	nodes := []struct {
		ID NodeID
	}{
		{1},
		{10},
		{6},
		{9},
		{12},
		{15},
	}

	for _, n := range nodes {
		table.Insert(RoutingEntry{n.ID, ""})
	}

	tests := []struct {
		Target     NodeID
		Candidates []NodeID
	}{
		{3, []NodeID{1}},
		{6, []NodeID{1}},
		{8, []NodeID{10}},
		{11, []NodeID{10}},
		{14, []NodeID{15}},
	}

	for _, tt := range tests {
		candidates := table.FindNode(tt.Target)
		for _, cc := range tt.Candidates {
			found := false
			for _, c := range candidates {
				if cc == c.NodeID {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("wrong candidate, should contain %d", cc)
			}
		}
	}
}

func TestRoutingTable_Big(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	for k := 1; k <= 100; k++ {
		owner := genNodeID()
		table := NewRoutingTable(k, owner, 1<<64-1)

		for i := 0; i < 1000; i++ {
			table.Insert(RoutingEntry{genNodeID(), ""})
		}
	}
}

func genNodeID() NodeID {
	return NodeID(uint64(rand.Uint32())<<32 + uint64(rand.Uint32()))
}
