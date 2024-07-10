package p2p

import (
	"github.com/souravbiswassanto/bit-torrent-client/client"
	"github.com/souravbiswassanto/bit-torrent-client/peers"
	"log"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	data  []byte
}

func (t *Torrent) Download() error {
	pieceStream := make(chan *pieceWork, len(t.PieceHashes))
	resultStream := make(chan *pieceResult)

	for index, piece := range t.PieceHashes {
		length := t.calculatePieceLength(index)
		pieceStream <- &pieceWork{index, piece, length}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, pieceStream, resultStream)
	}

	return nil
}

func (t *Torrent) calculatePieceLength(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, pieceStream chan *pieceWork, resultStream chan *pieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)
	c.SendUnchoke()
	c.SendInterested()

	for pw := range pieceStream {
		if !c.Bitfield.HasPiece(pw.index) {
			pieceStream <- pw
		}
	}
}
