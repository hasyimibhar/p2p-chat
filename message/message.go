package message

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"reflect"
)

const (
	NonceSize = 24
)

// Message is the interface that any message must implement.
type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) (Message, error)
}

// MessageFromOpcode returns the message type associated to the opcode.
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

// OpcodeFromMessage returns the opcode associated to the message type.
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
// - remaining bytes: the message body
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

	msgbuf = append([]byte{byte(opcode)}, msgbuf...)

	// Generate nonce
	nonce := make([]byte, NonceSize)
	if suite != nil {
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}

		msgbuf = suite.Seal(nil, nonce, msgbuf, nil)
	}

	encoded := make([]byte, len(msgbuf)+NonceSize)
	copy(encoded, nonce)
	copy(encoded[NonceSize:], msgbuf)

	return encoded, nil
}

// Decode decodes the byte slice into a message.
func Decode(buf []byte, suite cipher.AEAD, pubkey []byte) (Opcode, Message, error) {
	nonce := buf[:NonceSize]
	encrypted := buf[NonceSize:]

	var msgbuf []byte
	var err error

	// Decrypt message body
	if suite != nil {
		msgbuf, err = suite.Open(nil, nonce, encrypted, nil)
		if err != nil {
			return OpcodeNull, nil, err
		}
	} else {
		msgbuf = encrypted
	}

	opcode := Opcode(msgbuf[0])
	msg, err := MessageFromOpcode(opcode)
	if err != nil {
		return OpcodeNull, nil, err
	}

	msg, err = msg.Decode(msgbuf[1:])
	if err != nil {
		return OpcodeNull, nil, err
	}

	return opcode, msg, nil
}
