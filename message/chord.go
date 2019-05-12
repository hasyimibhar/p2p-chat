package message

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
