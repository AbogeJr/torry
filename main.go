package main

import (
	"fmt"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrentInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string             `bencode:"announce"`
	Info     bencodeTorrentInfo `bencode:"info"`
}

func (btfo *bencodeTorrent) toProcessedTorrentFile() (TorrentFile, error) {
	t := TorrentFile{
		Announce: btfo.Announce,
		// InfoHash: [],
		// PieceHashes: [],
		PieceLength: btfo.Info.PieceLength,
		Length:      btfo.Info.Length,
		Name:        btfo.Info.Name,
	}

	return t, nil
}

func openTorrentFile(filePath string) (TorrentFile, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("Error opening torrent file", err)
	}

	defer file.Close()

	bt := bencodeTorrent{}

	err = bencode.Unmarshal(file, &bt)

	if err != nil {
		fmt.Println("Error Unmarshalling:", err)
	}

	return bt.toProcessedTorrentFile()
}

func main() {
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	torrentFile, err := openTorrentFile(inputPath)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(outputPath, torrentFile)

	fmt.Printf("%+v\n", torrentFile)
}
