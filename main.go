package main

import (
	"fmt"
	"log"
	"os"

	tf "github.com/souravbiswassanto/bit-torrent-client/torrentfile"
)

func main() {
	args1 := os.Args[1]
	args2 := os.Args[2]
	fmt.Println(args1, args2)
	torrent, err := tf.Open(args1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(torrent.Announce, torrent.Name, torrent.Length)
	err = torrent.Download(args2)
	log.Fatalf("%v", err)
}
