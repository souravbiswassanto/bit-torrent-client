package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

// Port to listen on
const Port uint16 = 6881

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Name        string `bencode:"name"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func Open(filePath string) (TorrentFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TorrentFile{}, err
	}
	bt := bencodeTorrent{}

	err = bencode.Unmarshal(file, &bt)
	if err != nil {
		return TorrentFile{}, err
	}
	return bt.toTorrentFile()
}

func (t *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := t.Info.hash()
	if err != nil {
		return TorrentFile{}, nil
	}
	pieceHashes, err := t.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	return TorrentFile{
		Announce:    t.Announce,
		Length:      t.Info.Length,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
		PieceLength: t.Info.PieceLength,
		Name:        t.Info.Name,
	}, nil

}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	sz := 20
	buf := []byte(i.Pieces)
	if len(buf)%20 != 0 {
		return [][20]byte{}, fmt.Errorf("pieces corrupted")
	}
	tp := len(buf) / sz
	sph := make([][20]byte, tp)

	for i := 0; i < tp; i++ {
		copy(sph[i][:], buf[i*sz:(i+1)*sz])
	}
	return sph, nil
}

func (t *TorrentFile) Download(writePath string) error {
	var peerId [20]byte
	_, err := rand.Read(peerId[:])
	if err != nil {
		return err
	}
	peers, err := t.requestPeers(peerId, Port)
	if err != nil {
		return err
	}
	fmt.Println(len(peers),peers)
	// fmt.Print(peers)

	return nil
}
