package message

import (
	"github.com/hasyimibhar/p2p-chat/ed25519"
)

type Handshake struct {
	PublicKey []byte
	Addr      string
	Signature []byte
}

func NewHandshake(privkey []byte, pubkey []byte, addr string) (Handshake, error) {
	payload := append(pubkey, []byte(addr)...)

	sig, err := ed25519.Sign(privkey, pubkey, payload)
	if err != nil {
		return Handshake{}, err
	}

	return Handshake{
		PublicKey: pubkey,
		Addr:      addr,
		Signature: sig,
	}, nil
}

func (m Handshake) Verify() error {
	payload := append(m.PublicKey, []byte(m.Addr)...)
	return ed25519.Verify(m.PublicKey, payload, m.Signature)
}

func (m Handshake) Encode() ([]byte, error) {
	return append(m.PublicKey, append([]byte(m.Addr), m.Signature...)...), nil
}

func (m Handshake) Decode(buf []byte) (Message, error) {
	return Handshake{
		PublicKey: buf[:32],
		Addr:      string(buf[32 : len(buf)-64]),
		Signature: buf[len(buf)-64:],
	}, nil
}
