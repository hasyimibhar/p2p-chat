package message

import (
	"bytes"
	"testing"

	"github.com/hasyimibhar/p2p-chat/ed25519"
)

func TestChat_EncodeDecode(t *testing.T) {
	_, pub, _ := ed25519.GenerateKey()
	msg := Chat{
		PublicKey: pub,
		Text:      "lorem ipsum dolor sit amet",
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, append(pub, []byte("lorem ipsum dolor sit amet")...)) {
		t.Fatal("encoded message is incorrect")
	}

	decoded, err := Chat{}.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	chat, ok := decoded.(Chat)
	if !ok {
		t.Fatal("wrong message type")
	}

	if !bytes.Equal(chat.PublicKey, pub) {
		t.Fatal("decoded message is incorrect")
	}
	if chat.Text != "lorem ipsum dolor sit amet" {
		t.Fatal("decoded message is incorrect")
	}
}
