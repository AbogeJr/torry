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

	tf, err := torrentfile.OpenTorrentFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	progressChan := make(chan float64)
	buffChan := make(chan []byte)

	go func() {
		for p := range progressChan {
			fmt.Printf("\rPROGRESS: %.2f%%", p)
		}
	}()

	// TODO: eventually write the bytes to disk and clear the in-mem buffer
	// go func() {
	// 	for b := range buffChan {
	// 		fmt.Printf("\rBuffer%d", b[0])
	// 	}
	// }()

	err = tf.D2f(&progressChan, &buffChan)
	if err != nil {
		log.Fatal(err)
	}
}
