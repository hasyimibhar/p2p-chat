package main

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hasyimibhar/p2p-chat/ed25519"
	"github.com/hasyimibhar/p2p-chat/message"
)

const (
	// SuccessorListSize is the number of successors each
	// node keeps in its successor list (not including its
	// immediate successor).
	SuccessorListSize = 2
)

type chatEntry struct {
	PublicKey []byte
	Text      string
}

type Node struct {
	pubkey  []byte
	privkey []byte
	port    int

	ln          net.Listener
	mtx         sync.Mutex
	successor   *Peer
	successors  []string
	predecessor string
	suites      map[string]cipher.AEAD
	chatLog     []chatEntry
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
		successors:  make([]string, SuccessorListSize),
		predecessor: fmt.Sprintf("localhost:%d", port), // Set predecessor to self
		suites:      map[string]cipher.AEAD{},
		chatLog:     []chatEntry{},
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

	n.mtx.Lock()
	n.ln = ln
	n.mtx.Unlock()

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		// log.Printf("[trace] peer connected at %s", conn.RemoteAddr().String())

		peer := NewPeer(n, conn)
		if err := n.performHandshake(peer); err != nil {
			log.Println("[error] woot", err)
		}

		go n.handleMessages(peer)
	}
}

func (n *Node) Chat(text string) error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}

	if err := n.Successor().SendMessage(message.Chat{
		PublicKey: n.pubkey,
		Text:      text,
	}); err != nil {
		return err
	}

	n.mtx.Lock()
	n.chatLog = append(n.chatLog, chatEntry{
		Text:      text,
		PublicKey: n.pubkey,
	})
	n.mtx.Unlock()

	return nil
}

func (n *Node) StartPrivateChat(publicKey []byte) error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}

	return n.Successor().SendMessage(message.StartPrivateChatRequest{
		PublicKey: publicKey,
		Sender:    n.Addr(),
	})
}

func (n *Node) PrivateChat(publicKey []byte, text string) error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}
	suite, ok := n.suites[base64.StdEncoding.EncodeToString(publicKey)]
	if !ok {
		return fmt.Errorf("private chat has not been initialized with %s",
			base64.StdEncoding.EncodeToString(publicKey))
	}

	msg, err := message.NewPrivateChat(n.pubkey, publicKey, text, suite)
	if err != nil {
		return fmt.Errorf("failed to create private chat message: %s", err)
	}

	return n.Successor().SendMessage(msg)
}

func (n *Node) Close() {
	log.Println("[info] shutting down node")

	n.mtx.Lock()
	defer n.mtx.Unlock()

	n.ln.Close()
}

// connectToPeer connects to a peer and perform cryptographic handshake.
func (n *Node) connectToPeer(address string) (*Peer, error) {
	// log.Println("[trace] connecting to peer", address)

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

// JoinPeer makes the node to join the peer network
// and set the peer at the specified address as its successor.
func (n *Node) JoinPeer(address string) error {
	peer, err := n.connectToPeer(address)
	if err != nil {
		return err
	}

	if err := n.notify(peer); err != nil {
		return err
	}

	n.mtx.Lock()
	chatLogEmpty := len(n.chatLog) == 0
	n.mtx.Unlock()

	// Request chat log from peer if it's chat log is empty
	if chatLogEmpty {
		if err := peer.SendMessage(message.ChatLogRequest{}); err != nil {
			return err
		}
	}

	if n.Successor() == nil {
		go n.handleStabilize()
	}

	n.mtx.Lock()
	n.successor = peer
	n.mtx.Unlock()

	return nil
}

func (n *Node) performHandshake(peer *Peer) error {
	n.mtx.Lock()
	defer n.mtx.Unlock()

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

	// log.Printf("[trace] cryptographic handshake with peer %s successful", peer.Addr())
	return nil
}

func (n *Node) handleMessages(peer *Peer) {
	for {
		select {
		case msg := <-peer.ReceiveMessage(message.OpcodeChat):
			chat := msg.(message.Chat)

			n.mtx.Lock()
			n.chatLog = append(n.chatLog, chatEntry{
				Text:      chat.Text,
				PublicKey: chat.PublicKey,
			})
			n.mtx.Unlock()

			log.Printf("[%s] %s", base64.StdEncoding.EncodeToString(chat.PublicKey), chat.Text)

			// If the node's successor is not the sender of the chat message,
			// propagate the chat message to the successor, effectively
			// broadcasting the chat message.
			if n.Successor() != nil && !bytes.Equal(chat.PublicKey, n.Successor().PublicKey()) {
				if err := n.Successor().SendMessage(chat); err != nil {
					log.Println("[error] propagate chat failed:", err)
				}
			}

		case <-peer.ReceiveMessage(message.OpcodeChatLogRequest):
			msg := message.ChatLog{
				Entries: []message.Chat{},
			}

			n.mtx.Lock()
			for _, e := range n.chatLog {
				msg.Entries = append(msg.Entries, message.Chat{
					PublicKey: e.PublicKey,
					Text:      e.Text,
				})

				log.Printf("[%s] %s", base64.StdEncoding.EncodeToString(e.PublicKey), e.Text)
			}

			n.mtx.Unlock()

			if err := peer.SendMessage(msg); err != nil {
				log.Println("[error] chat log response failed:", err)
			}

		case msg := <-peer.ReceiveMessage(message.OpcodeChatLog):
			log := msg.(message.ChatLog)

			// TODO: Implement conflict resolution instead of
			// replacing the chat log
			n.mtx.Lock()
			n.chatLog = []chatEntry{}

			for _, e := range log.Entries {
				n.chatLog = append(n.chatLog, chatEntry{
					PublicKey: e.PublicKey,
					Text:      e.Text,
				})
			}

			n.mtx.Unlock()

		case msg := <-peer.ReceiveMessage(message.OpcodeNotify):
			n.rectify(peer, msg.(message.Notify))

		case <-peer.ReceiveMessage(message.OpcodeStabilizeRequest):
			n.mtx.Lock()
			err := peer.SendMessage(message.StabilizeResponse{
				Predecessor: n.predecessor,
			})
			if err != nil {
				log.Println("[error] stabilize response failed:", err)
			}
			n.mtx.Unlock()

		case msg := <-peer.ReceiveMessage(message.OpcodeStartPrivateChatRequest):
			info := msg.(message.StartPrivateChatRequest)

			if n.Successor() == nil {
				log.Println("[error] node has no successor")
			}

			// If the node is not the recipient of the message, pass it to its successor
			if !bytes.Equal(info.PublicKey, n.pubkey) {
				if err := n.Successor().SendMessage(info); err != nil {
					log.Println("[error] propagate message failed:", err)
				}
			} else if info.Sender == n.Addr() {
				log.Println("[error] recipient not found")
			} else {
				peer, err := n.connectToPeer(info.Sender)
				if err != nil {
					log.Println("[error] failed to connect to peer:", err)
				}

				n.suites[base64.StdEncoding.EncodeToString(peer.PublicKey())] = peer.CipherSuite()

				log.Println("[info] initialized private message with",
					base64.StdEncoding.EncodeToString(peer.PublicKey()))

				if err := peer.SendMessage(message.StartPrivateChatResponse{}); err != nil {
					log.Println("[error] failed to send message to peer:", err)
				} else {
					peer.Close()
				}
			}

		case <-peer.ReceiveMessage(message.OpcodeStartPrivateChatResponse):
			n.suites[base64.StdEncoding.EncodeToString(peer.PublicKey())] = peer.CipherSuite()
			log.Println("[info] initialized private message with",
				base64.StdEncoding.EncodeToString(peer.PublicKey()))

			peer.Close()

		case msg := <-peer.ReceiveMessage(message.OpcodePrivateChat):
			chat := msg.(message.PrivateChat)

			if n.Successor() == nil {
				log.Println("[error] node has no successor")
			}

			// If the node is not the recipient of the message, pass it to its successor
			if !bytes.Equal(chat.PublicKey, n.pubkey) {
				if err := n.Successor().SendMessage(chat); err != nil {
					log.Println("[error] propagate message failed:", err)
				}
			} else if bytes.Equal(chat.Sender, n.pubkey) {
				// The message has circled the whole network without finding
				log.Println("[error] recipient not found")
			} else {
				suite, ok := n.suites[base64.StdEncoding.EncodeToString(chat.Sender)]
				if !ok {
					log.Println("[error] cipher suite not found for peer", base64.StdEncoding.EncodeToString(chat.PublicKey))
				} else {
					text, err := chat.Decrypt(suite)
					if err != nil {
						log.Println("[error] decrypt private chat failed:", err)
					}

					log.Printf("[(private) %s] %s", base64.StdEncoding.EncodeToString(chat.Sender), text)
				}
			}

		case msg := <-peer.ReceiveMessage(message.OpcodeSuccessorRequest):
			if err := n.handleMessageSuccessorRequest(msg.(message.SuccessorRequest)); err != nil {
				log.Println("[error] propagate message failed:", err)
			}

		case msg := <-peer.ReceiveMessage(message.OpcodeSuccessorResponse):
			n.updateSuccessorList(msg.(message.SuccessorResponse))

		case msg := <-peer.ReceiveMessage(message.OpcodePing):
			if err := peer.SendMessage(msg); err != nil {
				log.Println("[error] ping failed:", err)
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
	// log.Printf("[trace] updating predecessor to %s", msg.Predecessor)

	n.mtx.Lock()
	n.predecessor = msg.Predecessor
	n.mtx.Unlock()

	// If a node has no successor, it means the node
	// is the initial node. If so, set the peer as its
	// successor and start the stabilization goroutine.
	if n.Successor() == nil {
		// log.Printf("[trace] updating successor to %s", peer.ListenAddr())
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
			if err := n.beginUpdateSuccessorList(); err != nil {
				log.Println("[error] populate successor list failed:", err)
			}
		}()
	}
}

func (n *Node) Successor() *Peer {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	return n.successor
}

func (n *Node) stabilize() error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}

	if err := n.Successor().SendMessage(message.Ping{}); err != nil {
		return err
	}

	select {
	case <-time.After(time.Second * 5):
		log.Printf("[warn] unable to contact peer %s, finding new successor from successor list", n.Successor().ListenAddr())
		n.Successor().Close()

		n.mtx.Lock()
		successors := make([]string, len(n.successors))
		copy(successors, n.successors)
		n.mtx.Unlock()

		for _, addr := range successors {
			if addr == "" {
				continue
			}

			if err := n.JoinPeer(addr); err == nil {
				// Found new successor
				log.Println("[info] found new successor:", addr)
				break
			}
		}

	case <-n.Successor().ReceiveMessage(message.OpcodePing):
	}

	// log.Printf("[trace] running periodic stabilize routine (successor=%s, predecessor=%s)",
	// 	n.Successor().ListenAddr(), n.predecessor)

	if err := n.Successor().SendMessage(message.StabilizeRequest{}); err != nil {
		return err
	}

	msg := <-n.Successor().ReceiveMessage(message.OpcodeStabilizeResponse)
	response := msg.(message.StabilizeResponse)

	if response.Predecessor == n.Addr() {
		return nil
	}

	// log.Printf("[trace] updating successor to %s", response.Predecessor)

	n.Successor().Close()

	if err := n.JoinPeer(response.Predecessor); err != nil {
		return err
	}

	return nil
}

func (n *Node) beginUpdateSuccessorList() error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}

	return n.Successor().SendMessage(message.SuccessorRequest{
		PublicKey: n.pubkey,
		Count:     0,
		Sender:    n.Addr(),
	})
}

func (n *Node) handleMessageSuccessorRequest(msg message.SuccessorRequest) error {
	if n.Successor() == nil {
		return fmt.Errorf("node has no successor")
	}

	// Don't propagate if the message have circled the network
	if bytes.Equal(n.Successor().PublicKey(), msg.PublicKey) {
		return nil
	}

	peer, err := n.connectToPeer(msg.Sender)
	if err != nil {
		return err
	}

	defer peer.Close()

	err = peer.SendMessage(message.SuccessorResponse{
		Count:     msg.Count,
		Successor: n.Successor().ListenAddr(),
	})
	if err != nil {
		return nil
	}

	// Propagate the message to the next successor
	msg.Count++
	if msg.Count < SuccessorListSize {
		return n.Successor().SendMessage(msg)
	}

	return nil
}

func (n *Node) updateSuccessorList(msg message.SuccessorResponse) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	n.successors[msg.Count] = msg.Successor
}
