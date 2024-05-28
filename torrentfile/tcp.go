package torrentfile

import (
	"net/http"
	"net/url"
	"time"

	bencode "github.com/jackpal/bencode-go"
	"github.com/souravbiswassanto/bit-torrent-client/peers"
)

func (t *TorrentFile) RequestPeersHTTP(tracker *url.URL, peerId [20]byte, port uint16) ([]peers.Peer, error) {

	httpClient := &http.Client{Timeout: time.Second * 15}
	// make a get request to the tracker for peers for this torrent file
	resp, err := httpClient.Get(tracker.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	trackerResponse := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResponse)
	if err != nil {
		return nil, err
	}
	return peers.Unmarshal([]byte(trackerResponse.Peers))
}
