package main

import (
	"fmt"
	"io"
	"strings"
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
	var test io.Reader = strings.NewReader(`d5:emailld5:where4:home4:addr15:gre@example.comed5:where4:work4:addr12:gre@work.comee4:name14:Grace R. Emlin7:address15:123 Main Streete`)
	var r = Result{}
	fmt.Println(test)
	
	err :=  bencode.Unmarshal(test, &r)
	if err != nil {
		fmt.Println("Error Unmarshalling:", err)
	}
	
	fmt.Printf("%+v\n", r)
}
