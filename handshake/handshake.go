package handshake

import "io"

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	idx := 1
	idx += copy(buf[idx:], h.Pstr)
	idx += copy(buf[idx:], make([]byte, 8))
	idx += copy(buf[idx:], h.InfoHash[:])
	idx += copy(buf[idx:], h.PeerID[:])
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	pstrLenBuf := make([]byte, 1)
	_, err := io.ReadFull(r, pstrLenBuf)
	if err != nil {
		return nil, err
	}
	pstrLen := int(pstrLenBuf[0])
	if pstrLen == 0 {
		return nil, err
	}

	response := make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, response)
	if err != nil {
		return nil, nil
	}
	pstr := string(response[:pstrLen])
	idx := pstrLen + 8
	infoHash := make([]byte, 20)
	peerID := make([]byte, 20)
	idx += copy(infoHash, response[idx:])
	idx += copy(peerID, response[idx:])
	return &Handshake{
		Pstr:     pstr,
		InfoHash: [20]byte(infoHash),
		PeerID:   [20]byte(peerID),
	}, nil
}
