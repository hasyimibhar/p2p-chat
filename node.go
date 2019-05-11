package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Node struct {
	pubkey  []byte
	privkey []byte
	port    int

	ln    net.Listener
	peers []*Peer
	mtx   sync.Mutex
}

func NewNode(port int) (*Node, error) {
	privkey, pubkey, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Node{
		pubkey:  pubkey,
		privkey: privkey,
		port:    port,
		peers:   []*Peer{},
	}, nil
}

func (n *Node) PublicKey() []byte  { return n.pubkey }
func (n *Node) PrivateKey() []byte { return n.privkey }

func (n *Node) ListenForConnections() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", n.port))
	if err != nil {
		return err
	}

	log.Println("[info] listenting for peers on", ln.Addr().String())
	n.ln = ln

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		log.Printf("[info] peer connected at %s", conn.RemoteAddr().String())

		peer := NewPeer(conn)

		n.mtx.Lock()
		n.peers = append(n.peers, peer)
		n.mtx.Unlock()

		go peer.Handle(n)
	}
}

func (n *Node) SendChatMessage(msg string) error {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	for _, p := range n.peers {
		if err := p.SendChatMessage(n, msg); err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) Close() {
	log.Println("[info] shutting down node")

	n.mtx.Lock()
	defer n.mtx.Unlock()

	n.ln.Close()

	for _, p := range n.peers {
		p.Close()
	}
}

func (n *Node) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	peer := NewPeer(conn)
	if err := peer.PerformHandshake(n.privkey, n.pubkey); err != nil {
		return err
	}

	log.Printf("[info] cryptographic handshake with peer %s successful", conn.RemoteAddr().String())

	n.mtx.Lock()
	n.peers = append(n.peers, peer)
	n.mtx.Unlock()

	go peer.Handle(n)
	return nil
}
