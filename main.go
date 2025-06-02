package main

import (
	"fmt"
	"log"
	"os"
	"torry/torrentfile"
	// tea "github.com/charmbracelet/bubbletea"
)

func main() {
	inputPath := os.Args[1]
	outputPath := os.Args[2]

	tf, err := torrentfile.OpenTorrentFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	progressChan := make(chan float64)

	go func() {
		for p := range progressChan {
			fmt.Printf("PROGRESS: %.2f%%\n", p)
		}
	}()

	err = tf.D2f(outputPath, progressChan)
	if err != nil {
		log.Fatal(err)
	}
}
