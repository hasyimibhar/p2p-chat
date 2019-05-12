package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hasyimibhar/p2p-chat/ed25519"
	"github.com/hasyimibhar/p2p-chat/message"
)

type Node struct {
	pubkey  []byte
	privkey []byte
	port    int

	ln          net.Listener
	mtx         sync.Mutex
	successor   *Peer
	predecessor string
}

// NewNode creates a new node.
func NewNode(port int) (*Node, error) {
	privkey, pubkey, err := ed25519.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Node{
		pubkey:      pubkey,
		privkey:     privkey,
		port:        port,
		predecessor: fmt.Sprintf("localhost:%d", port), // Set predecessor to self
	}, nil
}

func (n *Node) Addr() string       { return fmt.Sprintf("localhost:%d", n.port) }
func (n *Node) PublicKey() []byte  { return n.pubkey }
func (n *Node) PrivateKey() []byte { return n.privkey }

// ListenForConnections listens for peers.
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

		peer := NewPeer(n, conn)
		if err := n.performHandshake(peer); err != nil {
			log.Println("[error] woot", err)
		}

		go n.handleMessages(peer)
	}
}

func (n *Node) Chat(text string) error {
	if n.successor == nil {
		return fmt.Errorf("node has no successor")
	}

	return n.successor.SendMessage(message.Chat{
		PublicKey: n.pubkey,
		Text:      text,
	})
}

func (n *Node) Close() {
	log.Println("[info] shutting down node")

	n.mtx.Lock()
	defer n.mtx.Unlock()

	n.ln.Close()
}

// connectToPeer connects to a peer and perform cryptographic handshake.
func (n *Node) connectToPeer(address string) (*Peer, error) {
	log.Println("[info] connecting to peer", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	peer := NewPeer(n, conn)
	if err := n.performHandshake(peer); err != nil {
		return nil, err
	}

	go n.handleMessages(peer)

	return peer, nil
}

func (n *Node) JoinPeer(address string) error {
	peer, err := n.connectToPeer(address)
	if err != nil {
		return err
	}

	if err := n.notify(peer); err != nil {
		return err
	}

	if n.successor == nil {
		go n.handleStabilize()
	}

	n.mtx.Lock()
	n.successor = peer
	n.mtx.Unlock()

	return nil
}

func (n *Node) performHandshake(peer *Peer) error {
	request := message.Handshake{
		PublicKey: n.pubkey,
		Addr:      n.Addr(),
	}

	if err := peer.SendMessage(request); err != nil {
		return err
	}

	msg := <-peer.ReceiveMessage(message.OpcodeHandshake)
	handshake := msg.(message.Handshake)

	if err := peer.PerformHandshake(handshake.PublicKey, handshake.Addr); err != nil {
		return err
	}

	log.Printf("[info] cryptographic handshake with peer %s successful", peer.Addr())
	return nil
}

func (n *Node) handleMessages(peer *Peer) {
	for {
		select {
		case msg := <-peer.ReceiveMessage(message.OpcodeChat):
			chat := msg.(message.Chat)
			log.Printf("%s-> %s", peer.ListenAddr(), chat.Text)

			// If the node's successor is not the sender of the chat message,
			// propagate the chat message to the successor, effectively
			// broadcasting the chat message.
			if n.successor != nil && !bytes.Equal(chat.PublicKey, n.successor.PublicKey()) {
				if err := n.successor.SendMessage(chat); err != nil {
					log.Println("[error] propagate chat failed:", err)
				}
			}

		case msg := <-peer.ReceiveMessage(message.OpcodeNotify):
			n.rectify(peer, msg.(message.Notify))

		case <-peer.ReceiveMessage(message.OpcodeStabilizeRequest):
			err := peer.SendMessage(message.StabilizeResponse{
				Predecessor: n.predecessor,
			})
			if err != nil {
				log.Println("[error] stabilize response failed:", err)
			}

		}
	}
}

func (n *Node) notify(peer *Peer) error {
	return peer.SendMessage(message.Notify{
		Predecessor: n.Addr(),
	})
}

func (n *Node) rectify(peer *Peer, msg message.Notify) {
	log.Printf("[trace] updating predecessor to %s", msg.Predecessor)

	n.mtx.Lock()
	n.predecessor = msg.Predecessor
	n.mtx.Unlock()

	// If a node has no successor, it means the node
	// is the initial node. If so, set the peer as its
	// successor and start the stabilization goroutine.
	if n.successor == nil {
		log.Printf("[trace] updating successor to %s", peer.ListenAddr())
		if err := n.JoinPeer(peer.ListenAddr()); err != nil {
			log.Println("[error]", err)
		}
	}
}

func (n *Node) handleStabilize() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		go func() {
			if err := n.stabilize(); err != nil {
				log.Println("[error] stabilization failed:", err)
			}
		}()
	}
}

func (n *Node) stabilize() error {
	if n.successor == nil {
		return fmt.Errorf("node has no successor")
	}

	// log.Printf("[trace] running periodic stabilize routine (successor=%s, predecessor=%s)",
	// 	n.successor.ListenAddr(), n.predecessor)

	if err := n.successor.SendMessage(message.StabilizeRequest{}); err != nil {
		return err
	}

	msg := <-n.successor.ReceiveMessage(message.OpcodeStabilizeResponse)
	response := msg.(message.StabilizeResponse)

	if response.Predecessor == n.Addr() {
		return nil
	}

	log.Printf("[trace] updating successor to %s", response.Predecessor)

	n.successor.Close()

	if err := n.JoinPeer(response.Predecessor); err != nil {
		return err
	}

	return nil
}
