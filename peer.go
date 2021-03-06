package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/hasyimibhar/p2p-chat/ed25519"
	"github.com/hasyimibhar/p2p-chat/message"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const (
	SharedSecretSize = 32
)

// Peer represents another node in the network.
type Peer struct {
	node            *Node
	conn            net.Conn
	listenAddr      string
	pubkey          []byte
	secret          []byte
	suite           cipher.AEAD
	closed          bool
	closeCh         chan struct{}
	messageQueue    sync.Map
	mtx             sync.Mutex
	handshakeDoneCh chan struct{}
}

// NewPeer creates a peer.
func NewPeer(node *Node, conn net.Conn) *Peer {
	peer := &Peer{
		node:            node,
		conn:            conn,
		closeCh:         make(chan struct{}),
		handshakeDoneCh: make(chan struct{}),
	}

	go peer.handleReceive()

	return peer
}

func (p *Peer) Addr() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) ListenAddr() string {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	return p.listenAddr
}

func (p *Peer) PublicKey() []byte {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	return p.pubkey
}

func (p *Peer) CipherSuite() cipher.AEAD {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	return p.suite
}

// SendMessage sends a message to the peer.
func (p *Peer) SendMessage(msg message.Message) error {
	encoded, err := message.Encode(msg, p.CipherSuite(), p.node.PrivateKey(), p.node.PublicKey())
	if err != nil {
		return err
	}

	lenbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenbuf, uint32(len(encoded)))
	encoded = append(lenbuf, encoded...)

	// log.Println("[trace] sending:", hex.EncodeToString(encoded))

	_, err = p.conn.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}

// ReceiveMessage returns a channel which outputs messages with
// the specified opcode.
func (p *Peer) ReceiveMessage(opcode message.Opcode) <-chan message.Message {
	entry, _ := p.messageQueue.LoadOrStore(opcode, make(chan message.Message))
	return entry.(chan message.Message)
}

func (p *Peer) handleReceive() {
	defer close(p.closeCh)

	for {
		lenbuf := make([]byte, 4)
		_, err := p.conn.Read(lenbuf)
		if err == io.EOF {
			return
		}
		if err != nil {
			// log.Println("[error] failed to read from peer:", err)
			return
		}

		msgbuf := make([]byte, binary.BigEndian.Uint32(lenbuf))
		_, err = p.conn.Read(msgbuf)
		if err == io.EOF {
			return
		}
		if err != nil {
			// log.Println("[error] failed to read from peer:", err)
			return
		}

		// log.Println("[trace] received:", hex.EncodeToString(append(lenbuf, msgbuf...)))

		// TODO: there should be a better way to do this.
		// If nonce is non-zero, the message is encrypted, so decoding can only
		// be done after cryptographic handshake is complete.
		if messageEncrypted(msgbuf) {
			<-p.handshakeDoneCh
		}

		opcode, msg, err := message.Decode(msgbuf, p.CipherSuite(), p.PublicKey())
		if err != nil {
			log.Println("[error] failed to decode message:", err)
			return
		}

		entry, _ := p.messageQueue.LoadOrStore(opcode, make(chan message.Message))
		ch := entry.(chan message.Message)

		ch <- msg
	}
}

func messageEncrypted(msgbuf []byte) bool {
	nonce := msgbuf[:message.NonceSize]
	return !bytes.Equal(nonce, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

// PerformHandshake initializes the peer's AEAD cipher which
// completes the cryptographic handshake.
func (p *Peer) PerformHandshake(pubkey []byte, addr string) error {
	p.mtx.Lock()
	p.pubkey = pubkey
	p.listenAddr = addr
	p.mtx.Unlock()

	if err := p.initAEAD(); err != nil {
		return err
	}

	close(p.handshakeDoneCh)

	return nil
}

func (p *Peer) initAEAD() error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	ephemeralSecret, err := ed25519.ComputeSharedSecret(p.node.PrivateKey(), p.pubkey)
	if err != nil {
		return err
	}

	hash := sha256.New
	hkdf := hkdf.New(hash, ephemeralSecret, nil, nil)

	p.secret = make([]byte, SharedSecretSize)
	if _, err := hkdf.Read(p.secret); err != nil {
		return fmt.Errorf("failed to derive key")
	}

	p.suite, err = chacha20poly1305.NewX(p.secret)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) Close() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.closed = true
	p.conn.Close()
	<-p.closeCh
}
