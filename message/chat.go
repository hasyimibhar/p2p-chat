package message

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
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

// ChatLogRequest asks a peer for its chat log.
type ChatLogRequest struct{}

func (m ChatLogRequest) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (m ChatLogRequest) Decode(buf []byte) (Message, error) {
	return ChatLogRequest{}, nil
}

// ChatLog contains the public chat log. It's used for replicating
// the public chat log among all peers.
type ChatLog struct {
	Entries []Chat
}

func (m ChatLog) Encode() ([]byte, error) {
	encoded := make([]byte, 2)
	binary.BigEndian.PutUint16(encoded, uint16(len(m.Entries)))

	for _, c := range m.Entries {
		buflen := make([]byte, 4)
		ce, _ := c.Encode()

		binary.BigEndian.PutUint32(buflen, uint32(len(ce)))
		encoded = append(encoded, append(buflen, ce...)...)
	}

	return encoded, nil
}

func (m ChatLog) Decode(buf []byte) (Message, error) {
	decoded := ChatLog{
		Entries: []Chat{},
	}

	entriesLength := binary.BigEndian.Uint16(buf)
	buf = buf[2:]

	for i := uint16(0); i < entriesLength; i++ {
		buflen := binary.BigEndian.Uint32(buf)
		buf = buf[4:]

		entry, err := Chat{}.Decode(buf[:buflen])
		if err != nil {
			return nil, err
		}

		decoded.Entries = append(decoded.Entries, entry.(Chat))
		buf = buf[buflen:]
	}

	return decoded, nil
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
