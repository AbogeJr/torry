package message

import (
	"encoding/binary"
	"io"
)

type messageId uint8

const (
	MsgChoke         messageId = 0
	MsgUnchoke       messageId = 1
	MsgInterested    messageId = 2
	MsgNotInterested messageId = 3
	MsgHave          messageId = 4
	MsgBitfield      messageId = 5
	MsgRequest       messageId = 6
	MsgPiece         messageId = 7
	MsgCancel        messageId = 8
)

type Message struct {
	ID      messageId
	Payload []byte
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	// <length 4 bytes (uint32)><ID 1 byte><Payload len(m.Payload)>
	idAndPayloadLen := len(m.Payload) + 1
	buf := make([]byte, 4+idAndPayloadLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(idAndPayloadLen))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf
}

func Read(r io.Reader) (*Message, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)

	if length == 0 {
		return nil, nil
	}

	buf := make([]byte, length)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	ms := Message{
		ID:      messageId(buf[0]),
		Payload: buf[1:],
	}

	return &ms, nil
}
