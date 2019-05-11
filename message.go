package main

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
)

const (
	NonceSize = 12
)

const (
	OpcodeHandshake = 1
	OpcodeChat      = 10
)

type Message struct {
	Header    MessageHeader
	Body      []byte
	Signature []byte
}

type MessageHeader struct {
	Opcode byte
	Nonce  []byte
}

func (m *Message) Sign(privkey []byte, pubkey []byte) error {
	signature, err := Sign(privkey, pubkey, m.Encode())
	if err != nil {
		return err
	}

	m.Signature = signature
	return nil
}

func (m *Message) Verify(pubkey []byte) error {
	return Verify(pubkey, m.Encode(), m.Signature)
}

func (m *Message) Encode() []byte {
	encoded := make([]byte, 1+NonceSize+len(m.Body))
	encoded[0] = m.Header.Opcode

	if m.Header.Nonce != nil {
		copy(encoded[1:], m.Header.Nonce)
	}

	copy(encoded[1+NonceSize:], m.Body)

	return encoded
}

func (m *Message) Encrypt(aead cipher.AEAD) error {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	m.Header.Nonce = nonce
	m.Body = aead.Seal(nil, nonce, m.Body, nil)
	return nil
}

func (m *Message) Decrypt(aead cipher.AEAD) (err error) {
	m.Body, err = aead.Open(nil, m.Header.Nonce, m.Body, nil)
	return
}

func ReadMessage(rd io.Reader) (*Message, error) {
	lenbuf := make([]byte, 4)
	_, err := rd.Read(lenbuf)
	if err != nil {
		return nil, err
	}

	msglen := binary.BigEndian.Uint32(lenbuf)
	msg := make([]byte, msglen)
	_, err = rd.Read(msg)
	if err != nil {
		return nil, err
	}

	return &Message{
		Header: MessageHeader{
			Opcode: msg[0],
			Nonce:  msg[1 : NonceSize+1],
		},
		Body:      msg[NonceSize+1 : msglen-64],
		Signature: msg[msglen-64:],
	}, nil
}

func WriteMessage(w io.Writer, msg *Message) (err error) {
	msglen := 1 + NonceSize + len(msg.Body) + len(msg.Signature)

	msgbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgbuf, uint32(msglen))

	msgbuf = append(msgbuf, msg.Encode()...)
	msgbuf = append(msgbuf, msg.Signature...)

	_, err = w.Write(msgbuf)
	return
}
