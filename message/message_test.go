package message

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/hasyimibhar/p2p-chat/ed25519"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

func TestOpcodeFromMessage(t *testing.T) {
	tests := []struct {
		Message Message
		Opcode  Opcode
	}{
		{Handshake{}, OpcodeHandshake},
		{Chat{}, OpcodeChat},
		{Notify{}, OpcodeNotify},
		{StabilizeRequest{}, OpcodeStabilizeRequest},
		{StabilizeResponse{}, OpcodeStabilizeResponse},
		{&Handshake{}, OpcodeHandshake},
		{&Chat{}, OpcodeChat},
		{&Notify{}, OpcodeNotify},
		{&StabilizeRequest{}, OpcodeStabilizeRequest},
		{&StabilizeResponse{}, OpcodeStabilizeResponse},
	}

	for _, tt := range tests {
		opcode, err := OpcodeFromMessage(tt.Message)
		if err != nil {
			t.Fatal(err)
		}

		if opcode != tt.Opcode {
			t.Fatal("incorrect opcode from message")
		}
	}
}

type unregisteredMessage struct {
	Foobar string
}

func (m unregisteredMessage) Encode() ([]byte, error) {
	return []byte(m.Foobar), nil
}

func (m unregisteredMessage) Decode(buf []byte) (Message, error) {
	return unregisteredMessage{Foobar: string(buf)}, nil
}

func TestOpcodeFromMessage_Error(t *testing.T) {
	_, err := OpcodeFromMessage(unregisteredMessage{})
	if err.Error() != "invalid message type" {
		t.Fatal("unexpected error")
	}
}

func TestMessageFromOpcode(t *testing.T) {
	tests := []struct {
		Message Message
		Opcode  Opcode
	}{
		{Handshake{}, OpcodeHandshake},
		{Chat{}, OpcodeChat},
		{Notify{}, OpcodeNotify},
		{StabilizeRequest{}, OpcodeStabilizeRequest},
		{StabilizeResponse{}, OpcodeStabilizeResponse},
	}

	for _, tt := range tests {
		msg, err := MessageFromOpcode(tt.Opcode)
		if err != nil {
			t.Fatal(err)
		}

		if reflect.TypeOf(msg) != reflect.TypeOf(tt.Message) {
			t.Fatal("incorrect message from opcode")
		}
	}
}

func TestMessageFromOpcode_Error(t *testing.T) {
	_, err := MessageFromOpcode(Opcode(42))
	if err.Error() != "invalid opcode" {
		t.Fatal("unexpected error")
	}
}

func TestEncodeDecode(t *testing.T) {
	a, A, _ := ed25519.GenerateKey()
	b, B, _ := ed25519.GenerateKey()

	secretA, _ := ed25519.ComputeSharedSecret(a, B)
	secretB, _ := ed25519.ComputeSharedSecret(b, A)

	suiteA := cipherSuite(t, secretA)
	suiteB := cipherSuite(t, secretB)

	chatA := Chat{
		PublicKey: A,
		Text:      "lorem ipsum dolor sit amet",
	}

	encoded, err := Encode(chatA, suiteA, a, A)
	if err != nil {
		t.Fatal(err)
	}

	opcode, msg, err := Decode(encoded, suiteB, B)
	if err != nil {
		t.Fatal(err)
	}

	if opcode != OpcodeChat {
		t.Fatal("incorrect opcode")
	}

	chatB, ok := msg.(Chat)
	if !ok {
		t.Fatal("incorrect message type")
	}

	if !bytes.Equal(chatA.PublicKey, chatB.PublicKey) {
		t.Fatal("incorrect decoded message")
	}
	if chatA.Text != chatB.Text {
		t.Fatal("incorrect decoded message")
	}
}

func TestEncodeDecode_Notify(t *testing.T) {
	a, A, _ := ed25519.GenerateKey()
	b, B, _ := ed25519.GenerateKey()

	secretA, _ := ed25519.ComputeSharedSecret(a, B)
	secretB, _ := ed25519.ComputeSharedSecret(b, A)

	suiteA := cipherSuite(t, secretA)
	suiteB := cipherSuite(t, secretB)

	notifyA := Notify{
		Predecessor: "localhost:5432",
	}

	encoded, err := Encode(notifyA, suiteA, a, A)
	if err != nil {
		t.Fatal(err)
	}

	opcode, msg, err := Decode(encoded, suiteB, B)
	if err != nil {
		t.Fatal(err)
	}

	if opcode != OpcodeNotify {
		t.Fatal("incorrect opcode")
	}

	notifyB, ok := msg.(Notify)
	if !ok {
		t.Fatal("incorrect message type")
	}

	if notifyA.Predecessor != notifyB.Predecessor {
		t.Fatal("incorrect decoded message")
	}
}

func cipherSuite(t *testing.T, ephemeralSecret []byte) cipher.AEAD {
	hkdf := hkdf.New(sha256.New, ephemeralSecret, nil, nil)

	secret := make([]byte, 32)
	if _, err := hkdf.Read(secret); err != nil {
		t.Fatal(err)
	}

	suite, err := chacha20poly1305.NewX(secret)
	if err != nil {
		t.Fatal(err)
	}

	return suite
}
