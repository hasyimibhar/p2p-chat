package message

import (
	"crypto/cipher"
	"crypto/rand"
)

type Chat struct {
	PublicKey []byte
	Text      string
}

func (m Chat) Encode() ([]byte, error) {
	return append(m.PublicKey, []byte(m.Text)...), nil
}

func (m Chat) Decode(buf []byte) (Message, error) {
	return Chat{PublicKey: buf[:32], Text: string(buf[32:])}, nil
}

// StartPrivateChatRequest informs the peer with the specified public key
// that the node wants to start exchanging private messages.
type StartPrivateChatRequest struct {
	Sender    string
	PublicKey []byte
}

func (m StartPrivateChatRequest) Encode() ([]byte, error) {
	return append(m.PublicKey, []byte(m.Sender)...), nil
}

func (m StartPrivateChatRequest) Decode(buf []byte) (Message, error) {
	return StartPrivateChatRequest{PublicKey: buf[:32], Sender: string(buf[32:])}, nil
}

// StartPrivateChatResponse is a response of StartPrivateChatRequest.
type StartPrivateChatResponse struct{}

func (m StartPrivateChatResponse) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (m StartPrivateChatResponse) Decode(buf []byte) (Message, error) {
	return StartPrivateChatResponse{}, nil
}

type PrivateChat struct {
	Sender     []byte
	PublicKey  []byte
	Nonce      []byte
	Ciphertext []byte
}

func NewPrivateChat(nodePubkey []byte, pubkey []byte, text string, suite cipher.AEAD) (PrivateChat, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return PrivateChat{}, err
	}

	ciphertext := suite.Seal(nil, nonce, []byte(text), nil)
	return PrivateChat{Sender: nodePubkey, PublicKey: pubkey, Nonce: nonce, Ciphertext: ciphertext}, nil
}

func (m PrivateChat) Decrypt(suite cipher.AEAD) (string, error) {
	plaintext, err := suite.Open(nil, m.Nonce, m.Ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (m PrivateChat) Encode() ([]byte, error) {
	return append(m.Sender, append(m.PublicKey, append(m.Nonce, m.Ciphertext...)...)...), nil
}

func (m PrivateChat) Decode(buf []byte) (Message, error) {
	return PrivateChat{
		Sender:     buf[:32],
		PublicKey:  buf[32:64],
		Nonce:      buf[64:88],
		Ciphertext: buf[88:],
	}, nil
}
