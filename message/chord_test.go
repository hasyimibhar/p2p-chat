package message

import (
	"bytes"
	"testing"
)

func TestNotify_EncodeDecode(t *testing.T) {
	msg := Notify{
		Predecessor: "localhost:4321",
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, []byte("localhost:4321")) {
		t.Fatal("encoded message is incorrect")
	}

	decoded, err := Notify{}.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	notify, ok := decoded.(Notify)
	if !ok {
		t.Fatal("wrong message type")
	}

	if notify.Predecessor != "localhost:4321" {
		t.Fatal("decoded message is incorrect")
	}
}

func TestStabilizeRequest_EncodeDecode(t *testing.T) {
	msg := StabilizeRequest{}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, []byte{}) {
		t.Fatal("encoded message is incorrect")
	}

	decoded, err := StabilizeRequest{}.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := decoded.(StabilizeRequest)
	if !ok {
		t.Fatal("wrong message type")
	}
}

func TestStabilizeResponse_EncodeDecode(t *testing.T) {
	msg := StabilizeResponse{
		Predecessor: "localhost:4321",
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, []byte("localhost:4321")) {
		t.Fatal("encoded message is incorrect")
	}

	decoded, err := StabilizeResponse{}.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	notify, ok := decoded.(StabilizeResponse)
	if !ok {
		t.Fatal("wrong message type")
	}

	if notify.Predecessor != "localhost:4321" {
		t.Fatal("decoded message is incorrect")
	}
}
