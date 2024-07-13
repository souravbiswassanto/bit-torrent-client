package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/souravbiswassanto/bit-torrent-client/client"
	"github.com/souravbiswassanto/bit-torrent-client/message"
	"github.com/souravbiswassanto/bit-torrent-client/peers"
	"log"
	"time"
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

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
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
		buf, err := attemptToDownloadPieces(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			pieceStream <- pw
			return
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			return
		}
		c.SendHave(pw.index)
		resultStream <- &pieceResult{
			index: pw.index,
			data:  buf,
		}
	}
}

func attemptToDownloadPieces(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}
	// Setting a deadline helps get unresponsive peers unstuck.
	c.Conn.SetDeadline(time.Time{}.Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize

				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}
				err := c.SendRequest(state.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

func (pp *pieceProgress) readMessage() error {
	msg, err := pp.client.Read() // this call blocks

	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		pp.client.Choked = false
	case message.MsgChoke:
		pp.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		pp.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(pp.index, pp.buf, msg)
		if err != nil {
			return err
		}
		pp.downloaded += n
		pp.backlog--
	}
	return nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}
