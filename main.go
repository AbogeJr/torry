package main

import (
	"fmt"
	"os"
	bencode "github.com/jackpal/bencode-go"
)

type Email struct {
	Where string
	Addr string
}

type Result struct {
	Name string
	Phone string
	Email []Email
}

func main(){
	inputPath := os.Args[1]
	outputPath := os.Args[2]
	var r = Result{}
	fmt.Println(inputPath, outputPath)

	file, err := os.Open(inputPath)

	if err != nil {
		fmt.Println("Error opening torrent file", err)
	}
	defer file.Close()
	
	err =  bencode.Unmarshal(file, &r)
	if err != nil {
		fmt.Println("Error Unmarshalling:", err)
	}
	
	fmt.Printf("%+v\n", r)
}
