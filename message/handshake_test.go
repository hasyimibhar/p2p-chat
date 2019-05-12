package message

import (
	"bytes"
	"testing"

	"github.com/hasyimibhar/p2p-chat/ed25519"
)

func TestHandshake_EncodeDecodeVerify(t *testing.T) {
	priv, pub, _ := ed25519.GenerateKey()
	msg, _ := NewHandshake(priv, pub, "localhost:1234")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	sig, _ := ed25519.Sign(priv, pub, append(pub, []byte("localhost:1234")...))

	if !bytes.Equal(encoded, append(pub, append([]byte("localhost:1234"), sig...)...)) {
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

	if err := handshake.Verify(); err != nil {
		t.Fatal("message signature is invalid")
	}

	if !bytes.Equal(handshake.PublicKey, pub) {
		t.Fatal("decoded message is incorrect")
	}
	if handshake.Addr != "localhost:1234" {
		t.Fatal("decoded message is incorrect")
	}
	if !bytes.Equal(handshake.Signature, sig) {
		t.Fatal("decoded message is incorrect")
	}
}
