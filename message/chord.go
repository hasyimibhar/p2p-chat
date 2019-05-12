package message

import (
	"encoding/binary"
)

// Ping pings the peer.
type Ping struct{}

func (m Ping) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (m Ping) Decode(buf []byte) (Message, error) {
	return Ping{}, nil
}

// Notify notifies a peer to set the caller's
// node as its predecessor.
type Notify struct {
	Predecessor string
}

func (m Notify) Encode() ([]byte, error) {
	return []byte(m.Predecessor), nil
}

func (m Notify) Decode(buf []byte) (Message, error) {
	return Notify{Predecessor: string(buf)}, nil
}

// StabilizeRequest asks a peer for its predecessor.
type StabilizeRequest struct{}

func (m StabilizeRequest) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (m StabilizeRequest) Decode(buf []byte) (Message, error) {
	return StabilizeRequest{}, nil
}

// StabilizeResponse is the response of StabilizeRequest.
type StabilizeResponse struct {
	Predecessor string
}

func (m StabilizeResponse) Encode() ([]byte, error) {
	return []byte(m.Predecessor), nil
}

func (m StabilizeResponse) Decode(buf []byte) (Message, error) {
	return StabilizeResponse{Predecessor: string(buf)}, nil
}

// SuccessorRequest is sent by a peer to request another peer
// to return its successor. It's used to populate a peer's
// successor list.
type SuccessorRequest struct {
	Count     int
	PublicKey []byte
	Sender    string
}

func (m SuccessorRequest) Encode() ([]byte, error) {
	encoded := make([]byte, 4)
	binary.BigEndian.PutUint32(encoded, uint32(m.Count))

	return append(encoded, append(m.PublicKey, []byte(m.Sender)...)...), nil
}

func (m SuccessorRequest) Decode(buf []byte) (Message, error) {
	return SuccessorRequest{
		Count:     int(binary.BigEndian.Uint32(buf[:4])),
		PublicKey: buf[4:36],
		Sender:    string(buf[36:]),
	}, nil
}

// SuccessorResponse is the response of SuccessorRequest.
type SuccessorResponse struct {
	Count     int
	Successor string
}

func (m SuccessorResponse) Encode() ([]byte, error) {
	encoded := make([]byte, 4)
	binary.BigEndian.PutUint32(encoded, uint32(m.Count))

	return append(encoded, []byte(m.Successor)...), nil
}

func (m SuccessorResponse) Decode(buf []byte) (Message, error) {
	return SuccessorResponse{
		Count:     int(binary.BigEndian.Uint32(buf[:4])),
		Successor: string(buf[4:]),
	}, nil
}
