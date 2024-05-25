package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// requestPeers this requests the available peers from tracker.
// it first builds tracker urls and then make a http get request
// to the tracker. Tracker then returns the peer list in the response's body
func (t *TorrentFile) requestPeers(peerId [20]byte, port uint16) ([]string, error) {
	trackerUrl, err := t.buildTrackerUrls(peerId, port)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Timeout: time.Second * 15}
	// make a get request to the tracker for peers for this torrent file
	resp, err := httpClient.Get(trackerUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	trackerResponse := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResponse)
	return []string{trackerUrl}, err
}

// buildTrackerUrls builds a tracker urls from announce part of the
// TorrentFile. It lets tracker to know which file we want and announce
// our presence in the peerlist by queries params part
func (t *TorrentFile) buildTrackerUrls(peerId [20]byte, port uint16) (string, error) {
	tracker, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
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
	fmt.Println("Tracker url is: ", tracker.String())
	return tracker.String(), nil
}
