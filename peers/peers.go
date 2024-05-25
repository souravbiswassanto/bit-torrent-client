package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const PeerSize = 6

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Unmarshal receives a byte array, which is divideable by 6
// Each part of 6 bytes consists an IP(first 4 bytes) and an port (last 2 bytes) in big endian format
func Unmarshal(peerList []byte) ([]Peer, error) {
	if len(peerList)%PeerSize != 0 {
		return nil, fmt.Errorf("malformed peerList")
	}
	totalPeers := len(peerList) / PeerSize
	peers := make([]Peer, totalPeers)
	for i := 0; i < totalPeers; i++ {
		peers[i].IP = peerList[i*PeerSize : i*PeerSize+4]
		peers[i].Port = binary.BigEndian.Uint16(peerList[i*PeerSize+4 : i*PeerSize+6])
	}
	return peers, nil
}

func (p *Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
