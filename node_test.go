package main

import (
	"testing"
)

func TestNode_Pair(t *testing.T) {
	// TODO: Find free port instead of hardcoding
	node1, err := NewNode(8001)
	if err != nil {
		t.Fatal(err)
	}

	go node1.ListenForConnections()
	defer node1.Close()

	node2, err := NewNode(8002)
	if err != nil {
		t.Fatal(err)
	}

	go node2.ListenForConnections()
	defer node2.Close()

	err = node2.JoinPeer("localhost:8001")
	if err != nil {
		t.Fatal(err)
	}

	// Force node1 to stabilize to form the ring
	err = node1.Stabilize()
	if err != nil {
		t.Fatal(err)
	}

	err = node1.Chat("Hello, world")
	if err != nil {
		t.Fatal(err)
	}

	msg := <-node2.ChatMessages()
	if msg.Text != "Hello, world" {
		t.Fatal("incorrect message received")
	}

	err = node2.Chat("lorem ipsum dolor sit amet")
	if err != nil {
		t.Fatal(err)
	}

	msg = <-node1.ChatMessages()
	if msg.Text != "lorem ipsum dolor sit amet" {
		t.Fatal("incorrect message received")
	}
}
