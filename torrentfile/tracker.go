package torrentfile

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/souravbiswassanto/bit-torrent-client/peers"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// requestPeers this requests the available peers from tracker.
// it first builds tracker urls and then make a http get request
// to the tracker. Tracker then returns the peer list in the response's body
func (t *TorrentFile) requestPeers(peerId [20]byte, port uint16) ([]peers.Peer, error) {
	tracker, err := t.buildTrackerUrl(peerId, port)
	if err != nil {
		return nil, err
	}

	switch tracker.Scheme {
	case "http", "https":
		return t.RequestPeersHTTP(tracker, peerId, port)
	case "udp":
		return t.RequestPeersUDP(tracker, peerId, port)
	default:
		return nil, fmt.Errorf("unsupported protocol scheme")

	}
}

// buildTrackerUrls builds a tracker urls from announce part of the
// TorrentFile. It lets tracker to know which file we want and announce
// our presence in the peerlist by queries params part
func (t *TorrentFile) buildTrackerUrl(peerId [20]byte, port uint16) (*url.URL, error) {
	tracker, err := url.Parse(t.Announce)
	if err != nil {
		return nil, err
	}

	quries := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	tracker.RawQuery = quries.Encode()

	return tracker, nil
}

// func (t *TorrentFile) solve(tracker *url.URL) {
// 	server, err := net.ResolveUDPAddr("udp", tracker.Host)
// 	if err != nil {
// 		fmt.Printf("failed to resolve UDP address: %w", err)
// 		return
// 	}
// 	fmt.Println(server.String())
// 	conn, err := net.DialUDP("udp", nil, server)
// 	if err != nil {
// 		fmt.Printf("failed to dial UDP: %w", err)
// 		return
// 	}
// 	defer conn.Close()
// 	rn := rnd.Int31()
// 	_, err = conn.Write(buildConnectRequest(rn))
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	// Read the connect response
// 	response := make([]byte, 16)
// 	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
// 	_, err = conn.Read(response)
// 	if err != nil {
// 		fmt.Printf("failed to read connect response: %w", err)
// 	}
// 	fmt.Println(string(response))

// 	var bt []byte
// 	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
// 	_, err = conn.Read(bt)
// 	if err != nil {
// 		log.Fatalln("second ", err)
// 	}
// 	fmt.Println("dkdl ", string(bt))
// }
