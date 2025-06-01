package downloader

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"
	"torry/client"
	"torry/message"
	"torry/peers"
)

const MAX_BACKLOG = 5
const MAX_BLOCK_SIZE = 16384

type Torrent struct {
	Peers       []peers.Peer
	InfoHash    [20]byte
	PeerID      [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(time.Second * 30))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MAX_BACKLOG && state.requested < pw.length {
				blockSize := min(pw.length-state.requested, MAX_BLOCK_SIZE)

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}

				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)

	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("piece index [%d] failed integrity check", pw.index)
	}

	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID)
	if err != nil {
		fmt.Println(err.Error())
		log.Printf("Could not complete handshake. Disconnecting. peer: %s", peer.Stringify())
		return
	}

	defer c.Conn.Close()
	log.Printf("Completed Handshake. peer: %s\n", peer.Stringify())

	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw
			continue
		}

		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBounds(index int) (bagin int, end int) {
	begin := index * t.PieceLength
	end = min(begin+t.PieceLength, t.Length)

	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBounds(index)
	return end - begin
}

func (t *Torrent) Download(progressChan chan<- float64) ([]byte, error) {
	log.Println("Starting Download for", t.Name)

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0

	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBounds(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		progressChan <- float64(donePieces) / float64(len(t.PieceHashes)) * 100
		// numWorkers := runtime.NumGoroutine() - 1
		// log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}

	close(workQueue)

	return buf, nil
}
