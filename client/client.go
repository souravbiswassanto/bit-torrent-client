package client

import (
	"bytes"
	"fmt"
	"github.com/souravbiswassanto/bit-torrent-client/bitfield"
	"github.com/souravbiswassanto/bit-torrent-client/handshake"
	"github.com/souravbiswassanto/bit-torrent-client/message"
	"github.com/souravbiswassanto/bit-torrent-client/peers"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func New(peer peers.Peer, infoHash, peerID [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("udp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}
	bf, err := receiveBitField(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil

}
func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	defer conn.SetDeadline(time.Time{})

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}
	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infoHash)
	}
	return res, nil
}

func receiveBitField(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Expected bitfield but got %s", msg)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}
	return msg.Payload, nil
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendRequest(index, begin, length int) error {
	msg := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}
