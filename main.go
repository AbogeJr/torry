package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	handshake "torry/handshake"
	message "torry/message"
	peers "torry/peers"
	torrentfile "torry/torrentfile"

	bencode "github.com/jackpal/bencode-go"
)

func main() {
	inputPath := os.Args[1]
	// outputPath := os.Args[2]

	torrentFile, err := torrentfile.OpenTorrentFile(inputPath)

	if err != nil {
		fmt.Println(err)
	}

	var peerID [20]byte

	_, err = rand.Read(peerID[:])
	if err != nil {
		fmt.Println(err)
	}

	trackerURL, err := torrentFile.BuildTrackerURL(peerID, 2131)
	/* note
	This is what the URL looks like after encoding the params
		http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info
		_hash=%A5%94%DF%3A%B9%B4%D3%95b%CC%05%F91%98i%A2%18%12%1FG&left=67842
		8672&peer_id=e%F5_GM%B3hRJ~%A69%04%EC%9Ft%98%B3%14%24&port=2131&uploa
		ded=0
	*/
	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.Get(trackerURL)

	if err != nil {
		fmt.Println("Error sending request to tracker URL", err)
	}

	defer resp.Body.Close()

	trackerResponse := peers.TrackerURLResponse{}

	err = bencode.Unmarshal(resp.Body, &trackerResponse)

	if err != nil {
		fmt.Println("Error parsing tracker response", err)
	}

	/*
		note: casting the bencoded peers binary string to a slice of bytes for
		processing/unmarshalling
	*/
	// peers, err := UnmarshallPeers([]byte(trackerResponse.Peers))

	// if err != nil {
	// 	fmt.Println("Error unmarshalling peers", err)
	// }

	peerHandshake := handshake.Handshake{
		Pstr:     "Bittorrent Protocol",
		Infohash: torrentFile.InfoHash,
		PeerId:   peerID,
	}
	serializedHandshake := peerHandshake.Serialize()

	// fmt.Println(peers)
	// fmt.Println(serializedHandshake)
	reader := bytes.NewReader(serializedHandshake)
	hs, err := handshake.ReadHandshake(reader)
	if err != nil {
		fmt.Println("Error reading handshake")
	}
	fmt.Printf("%+v\n", *hs)

	fmt.Println("Message test")

	ms := message.Message{
		ID:      69,
		Payload: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	serializedMessage := ms.Serialize()
	fmt.Println("Serialized message", serializedMessage)
	reader = bytes.NewReader(serializedMessage)
	msg, err := message.Read(reader)
	if err != nil {
		fmt.Println("Error reading handshake", err)
	}
	fmt.Println("Read Message", msg)

}
