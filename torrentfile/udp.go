package torrentfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	rnd "math/rand"
	"net"
	"net/url"
	"time"

	"github.com/souravbiswassanto/bit-torrent-client/peers"
)

// requestPeersUDP requests the peer list from udp server
// http://bittorrent.org/beps/bep_0015.html this explains how to do it
func (t *TorrentFile) RequestPeersUDP(tracker *url.URL, peerID [20]byte, port uint16) ([]peers.Peer, error) {
	server, err := net.ResolveUDPAddr("udp", tracker.Host)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, server)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	trxID := rnd.Int31()
	connReq := buildConnectRequest(trxID)
	_, err = conn.Write(connReq)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 16)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, err = conn.Read(response)
	if err != nil {
		return nil, err
	}
	connectionID, err := parseConnectResponse(response, trxID)
	if err != nil {
		return nil, err
	}
	trxID = rnd.Int31()
	announceReq, err := t.buildAnnounceRequest(trxID, connectionID, peerID, port)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(announceReq)
	if err != nil {
		return nil, err
	}
	announceResponse := make([]byte, 4096)
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	n, err := conn.Read(announceResponse)
	if err != nil {
		return nil, err
	}
	peers, err := parseAnnounceResponse(announceResponse, trxID, n)
	return peers, err
}

func buildConnectRequest(transactionID int32) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int64(0x41727101980)) // Connection ID
	binary.Write(&buf, binary.BigEndian, int32(0))             // Action (0 for connect)
	binary.Write(&buf, binary.BigEndian, transactionID)        // Transaction ID
	fmt.Println(buf.Bytes())
	return buf.Bytes()
}

func (t *TorrentFile) buildAnnounceRequest(trxID int32, connID int64, peerID [20]byte, port uint16) ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, connID)
	binary.Write(&buf, binary.BigEndian, int32(1))
	binary.Write(&buf, binary.BigEndian, int32(trxID))
	_, err := buf.Write(t.InfoHash[:])
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(peerID[:])
	if err != nil {
		return nil, err
	}
	binary.Write(&buf, binary.BigEndian, int64(0))
	binary.Write(&buf, binary.BigEndian, int64(t.Length))
	binary.Write(&buf, binary.BigEndian, int64(0))
	binary.Write(&buf, binary.BigEndian, int32(0))
	binary.Write(&buf, binary.BigEndian, int32(0))
	binary.Write(&buf, binary.BigEndian, (rnd.Int31()))
	binary.Write(&buf, binary.BigEndian, int32(-1)) // Num want: -1 (default)
	binary.Write(&buf, binary.BigEndian, port)
	return buf.Bytes(), nil
}

func parseConnectResponse(response []byte, trxID int32) (int64, error) {
	if len(response) < 16 {
		return 0, fmt.Errorf("connect response too short")
	}
	action := binary.BigEndian.Uint32(response[0:4])
	resTrxID := binary.BigEndian.Uint32(response[4:8])
	if action != 0 || int32(resTrxID) != trxID {
		return 0, fmt.Errorf("invalid connect response")
	}

	connID := binary.BigEndian.Uint64(response[8:16])
	return int64(connID), nil
}

func parseAnnounceResponse(response []byte, trxID int32, byteRead int) ([]peers.Peer, error) {
	if len(response) < 20 {
		return nil, fmt.Errorf("malformed announce response. response should be alteast 20 bytes")
	}
	action := binary.BigEndian.Uint32(response[0:4])
	resTrxID := binary.BigEndian.Uint32(response[4:8])
	if action != 1 || int32(resTrxID) != trxID {
		return nil, fmt.Errorf("invalid announce response")
	}
	return peers.Unmarshal(response[20:byteRead])
}
