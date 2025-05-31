package client

import (
	"bytes"
	"fmt"
	"net"
	"time"
	"torry/bitfield"
	"torry/handshake"
	"torry/message"
	"torry/peers"
)

type Client struct {
	Conn     net.Conn
	Peer     peers.Peer
	Bitfield bitfield.Bitfield
	Choked   bool
	InfoHash [20]byte
	PeerID   [20]byte
}

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}
func receiveBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected messageID: %d but goe messageID: %d", message.MsgBitfield, msg.ID)
	}

	return msg.Payload, nil
}

func New(peer peers.Peer, infohash [20]byte, peerID [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.Stringify(), time.Second*15)

	if err != nil {
		return nil, err
	}

	_, err = completeHandshake(conn, infohash, peerID)

	if err != nil {
		conn.Close()
		return nil, err
	}

	bitfield, err := receiveBitfield(conn)

	if err != nil {
		conn.Close()
		return nil, err
	}

	client := Client{
		Conn:     conn,
		Peer:     peer,
		Bitfield: bitfield,
		Choked:   true,
		InfoHash: infohash,
		PeerID:   peerID,
	}

	return &client, nil
}

func (client Client) Read() (*message.Message, error) {
	msg, err := message.Read(client.Conn)
	return msg, err
}

func (client Client) SendRequest(index, begin, length int) error {
	msg := message.FormatRequestMsg(index, begin, length)

	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}

	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}

	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}

	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client Client) SendHave(index int) error {
	msg := message.FormatHaveMsg(index)

	_, err := client.Conn.Write(msg.Serialize())
	return err
}
