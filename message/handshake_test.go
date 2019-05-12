package message

import (
	"bytes"
	"testing"

	"github.com/hasyimibhar/p2p-chat/ed25519"
)

func TestHandshake_EncodeDecode(t *testing.T) {
	_, pub, _ := ed25519.GenerateKey()
	msg := Handshake{
		PublicKey: pub,
		Addr:      "localhost:1234",
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, append(pub, []byte("localhost:1234")...)) {
		t.Fatal("encoded message is incorrect")
	}

	decoded, err := Handshake{}.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	handshake, ok := decoded.(Handshake)
	if !ok {
		t.Fatal("wrong message type")
	}

	if !bytes.Equal(handshake.PublicKey, pub) {
		t.Fatal("decoded message is incorrect")
	}
	if handshake.Addr != "localhost:1234" {
		t.Fatal("decoded message is incorrect")
	}
}
