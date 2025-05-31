package main

import (
	"log"
	"os"
	"torry/torrentfile"
)

func main() {
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	tf, err := torrentfile.OpenTorrentFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.D2f(outputPath)
	if err != nil {
		log.Fatal(err)
	}

}
