package message

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
