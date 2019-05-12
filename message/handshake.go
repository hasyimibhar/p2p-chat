package message

type Handshake struct {
	PublicKey []byte
	Addr      string
}

func (m Handshake) Encode() ([]byte, error) {
	return append(m.PublicKey, []byte(m.Addr)...), nil
}

func (m Handshake) Decode(buf []byte) (Message, error) {
	return Handshake{PublicKey: buf[:32], Addr: string(buf[32:])}, nil
}
