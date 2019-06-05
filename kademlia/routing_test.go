package kademlia

import (
	"math/rand"
	"testing"
	"time"
)

func TestRoutingTable_Simple(t *testing.T) {
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
		inserted := table.Insert(n.ID)
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
			for _, nn := range allNodes {
				if nn == n.ID {
					return
				}
			}
			t.Fatal("node list is wrong")
		} else {
			for _, nn := range allNodes {
				if nn == n.ID {
					t.Fatal("node list is wrong")
				}
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
			table.Insert(genNodeID())
		}
	}
}

func genNodeID() NodeID {
	return NodeID(uint64(rand.Uint32())<<32 + uint64(rand.Uint32()))
}
