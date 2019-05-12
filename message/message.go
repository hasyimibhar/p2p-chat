package message

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"reflect"

	"github.com/hasyimibhar/p2p-chat/ed25519"
)

const (
	NonceSize = 24
)

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) (Message, error)
}

func MessageFromOpcode(opcode Opcode) (Message, error) {
	mtx.Lock()
	defer mtx.Unlock()

	tp, ok := opcodes[Opcode(opcode)]
	if !ok {
		return nil, fmt.Errorf("invalid opcode")
	}

	message, ok := reflect.New(reflect.TypeOf(tp)).Elem().Interface().(Message)
	if !ok {
		return nil, fmt.Errorf("invalid opcode")
	}

	return message, nil
}

func OpcodeFromMessage(msg Message) (Opcode, error) {
	mtx.Lock()
	defer mtx.Unlock()

	typ := reflect.TypeOf(msg)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	opcode, exists := messages[typ]
	if !exists {
		return OpcodeNull, fmt.Errorf("invalid message type")
	}

	return opcode, nil
}

// Encode encodes the message for transport, and at the same time
// encrypts and signs the message.
//
// The format is as follows:
//
// - 1 byte: message opcode
// - 24 bytes: message nonce (for AEAD)
// - remaining bytes - 64 bytes: the message body
// - 64 bytes: message signature
//
func Encode(msg Message, suite cipher.AEAD, privkey []byte, pubkey []byte) ([]byte, error) {
	opcode, err := OpcodeFromMessage(msg)
	if err != nil {
		return nil, err
	}

	msgbuf, err := msg.Encode()
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, NonceSize)
	if suite != nil {
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}

		msgbuf = suite.Seal(nil, nonce, msgbuf, nil)
	}

	encoded := make([]byte, len(msgbuf)+NonceSize+1)
	encoded[0] = byte(opcode)
	copy(encoded[1:], nonce)
	copy(encoded[1+NonceSize:], msgbuf)

	// Sign the message
	signature, err := ed25519.Sign(privkey, pubkey, encoded)
	if err != nil {
		return nil, err
	}

	// Append signature
	return append(encoded, signature...), nil
}

func Decode(buf []byte, suite cipher.AEAD, pubkey []byte) (Opcode, Message, error) {
	opcode := Opcode(buf[0])
	nonce := buf[1:25]
	encrypted := buf[25 : len(buf)-64]
	// sig := buf[len(buf)-64:]

	// Verify message signature
	// payload := buf[:len(buf)-64]
	// if err := ed25519.Verify(pubkey, payload, sig); err != nil {
	// 	return OpcodeNull, nil, err
	// }

	var body []byte
	var err error

	// Decrypt message body
	if suite != nil {
		body, err = suite.Open(nil, nonce, encrypted, nil)
		if err != nil {
			return OpcodeNull, nil, err
		}
	} else {
		body = encrypted
	}

	msg, err := MessageFromOpcode(opcode)
	if err != nil {
		return OpcodeNull, nil, err
	}

	msg, err = msg.Decode(body)
	if err != nil {
		return OpcodeNull, nil, err
	}

	return opcode, msg, nil
}
