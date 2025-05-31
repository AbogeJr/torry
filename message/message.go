package message

import (
	"encoding/binary"
	"fmt"
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

func FormatRequestMsg(index, begin, length int) *Message {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	m := Message{
		ID:      MsgRequest,
		Payload: payload,
	}

	return &m
}

func FormatHaveMsg(index int) *Message {
	payload := make([]byte, 4)

	binary.BigEndian.PutUint32(payload, uint32(index))

	m := Message{
		ID:      MsgHave,
		Payload: payload,
	}

	return &m
}

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected messageID %d but got %d", MsgPiece, msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("too short ! (that's what she said): [%d]", len(msg.Payload))
	}

	pieceIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))

	if pieceIndex != index {
		return 0, fmt.Errorf("expected piece index [%d] but got [%d]", index, pieceIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))

	if begin >= len(buf) {
		return 0, fmt.Errorf("beginning index [%d] is too large for buffer length [%d]", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data is [%d] bytes; too large for buffer length [%d]", begin, len(buf))
	}

	copy(buf[begin:], data)

	return len(data), nil
}

func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected messageID %d but got %d", MsgHave, msg.ID)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload to be 4bytes. got [%d]", len(msg.Payload))
	}

	index := int(binary.BigEndian.Uint32(msg.Payload))

	return index, nil
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

func (m *Message) name() string {
	if m == nil {
		return "KeepAlive"
	}

	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}

}

func (m *Message) Stringify() string {
	if m == nil {
		return m.name()
	}

	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}
