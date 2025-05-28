package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	Pstr     string
	Infohash [20]byte
	PeerId   [20]byte
}

func (h *Handshake) Serialize() []byte {
	// note: the bittorrent handshake will be 68 bytes long
	buf := make([]byte, len(h.Pstr)+49)

	buf[0] = byte(len(h.Pstr))
	currentIndex := 1 //note: buff[0] is currently holding our pstr length
	currentIndex += copy(buf[currentIndex:], []byte(h.Pstr))
	currentIndex += copy(buf[currentIndex:], make([]byte, 8))
	currentIndex += copy(buf[currentIndex:], h.Infohash[:])
	currentIndex += copy(buf[currentIndex:], h.PeerId[:])

	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {

	pstrLengthBuffer := make([]byte, 1)

	_, err := io.ReadFull(r, pstrLengthBuffer)

	if err != nil {
		return &Handshake{}, err
	}

	pstrlen := int(pstrLengthBuffer[0])

	if pstrlen == 0 {
		fmt.Println("pstr length cannot be 0")
	}

	handshakeBuff := make([]byte, pstrlen+48)
	_, err = io.ReadFull(r, handshakeBuff)
	if err != nil {
		return &Handshake{}, err
	}

	var peerId, infoHash [20]byte

	copy(infoHash[:], handshakeBuff[pstrlen+8:pstrlen+8+20])
	copy(peerId[:], handshakeBuff[pstrlen+8+20:pstrlen+8+20+20])

	h := Handshake{
		Pstr:     string(handshakeBuff[:pstrlen]),
		Infohash: infoHash,
		PeerId:   peerId,
	}

	return &h, nil
}
