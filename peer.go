package main

import (
	"crypto/cipher"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/chacha20poly1305"
)

type Peer struct {
	conn    net.Conn
	pubkey  []byte
	secret  []byte
	aead    cipher.AEAD
	closed  bool
	closeCh chan struct{}
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		conn:    conn,
		closeCh: make(chan struct{}),
	}
}

func (p *Peer) PublicKey() []byte {
	return p.pubkey
}

func (p *Peer) Handle(node *Node) {
	defer func() {
		if !p.closed {
			log.Printf("[info] lost connection with peer %s", p.conn.RemoteAddr().String())
		}
	}()

	defer close(p.closeCh)

	for {
		msg, err := ReadMessage(p.conn)
		if err != nil {
			if !p.closed && err != io.EOF {
				log.Println("[error] failed to read from peer", err)
			}

			return
		}

		if !p.Authenticated() && msg.Header.Opcode != OpcodeHandshake {
			log.Println("[error] received message from unauthenticated peer")
			return
		}

		switch msg.Header.Opcode {
		case OpcodeHandshake:
			if p.Authenticated() {
				log.Println("[error] received handshake message from authenticated peer")
				return
			}

			peerPubkey := msg.Body
			if err := msg.Verify(peerPubkey); err != nil {
				log.Println("[error]", err)
				return
			}

			if err := p.ReceiveHandshake(peerPubkey, node.PrivateKey(), node.PublicKey()); err != nil {
				log.Println("[error] failed to perform handshake with peer:", err)
				return
			}

			log.Printf("[info] cryptographic handshake with peer %s successful", p.conn.RemoteAddr().String())

		case OpcodeChat:
			if err := msg.Verify(p.pubkey); err != nil {
				log.Println("[error]", err)
				return
			}

			if err := msg.Decrypt(p.aead); err != nil {
				log.Println("[error]", err)
				return
			}

			log.Printf("%s-> %s", p.conn.RemoteAddr().String(), string(msg.Body))
		}
	}
}

func (p *Peer) PerformHandshake(nodePrivkey []byte, nodePubkey []byte) error {
	// Send pubkey to peer to start the handshake
	request := &Message{
		Header: MessageHeader{Opcode: OpcodeHandshake},
		Body:   nodePubkey,
	}

	var err error

	if err := request.Sign(nodePrivkey, nodePubkey); err != nil {
		return err
	}

	if err := WriteMessage(p.conn, request); err != nil {
		return err
	}

	response, err := ReadMessage(p.conn)
	if err != nil {
		return err
	}

	if response.Header.Opcode != OpcodeHandshake {
		return fmt.Errorf("unexpected opcode: expecting handshake")
	}

	peerPubkey := response.Body
	if err := response.Verify(peerPubkey); err != nil {
		return err
	}

	if err := p.initAEAD(nodePrivkey, peerPubkey); err != nil {
		return err
	}

	return nil
}

func (p *Peer) ReceiveHandshake(peerPubkey []byte, nodePrivkey []byte, nodePubkey []byte) error {
	// Send node pubkey to peer to complete the handshake
	request := &Message{
		Header: MessageHeader{Opcode: OpcodeHandshake},
		Body:   nodePubkey,
	}

	if err := request.Sign(nodePrivkey, nodePubkey); err != nil {
		return err
	}

	if err := WriteMessage(p.conn, request); err != nil {
		return err
	}

	if err := p.initAEAD(nodePrivkey, peerPubkey); err != nil {
		return err
	}

	return nil
}

func (p *Peer) initAEAD(nodePrivkey []byte, peerPubkey []byte) error {
	secret, err := ComputeSharedSecret(nodePrivkey, peerPubkey)
	if err != nil {
		return err
	}

	p.pubkey = peerPubkey
	p.secret = secret

	p.aead, err = chacha20poly1305.NewX(p.secret)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) SendChatMessage(node *Node, msg string) (err error) {
	request := &Message{
		Header: MessageHeader{Opcode: OpcodeChat},
		Body:   []byte(msg),
	}

	if err = request.Encrypt(p.aead); err != nil {
		err = fmt.Errorf("failed to encrypt message: %s", err)
		return
	}

	if err = request.Sign(node.PrivateKey(), node.PublicKey()); err != nil {
		err = fmt.Errorf("failed to sign message: %s", err)
		return
	}

	if err = WriteMessage(p.conn, request); err != nil {
		return
	}

	return
}

func (p *Peer) Authenticated() bool {
	return p.secret != nil
}

func (p *Peer) Close() {
	p.closed = true
	p.conn.Close()
	<-p.closeCh
}
