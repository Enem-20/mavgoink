package message

import (
	"encoding/binary"
	"errors"
	"math"
	"unsafe"
)

const (
	MAVLINK_MAX_PAYLOAD_LEN       = 255
	MAVLINK_CORE_HEADER_LEN       = 9
	MAVLINK_NUM_HEADER_BYTES      = MAVLINK_CORE_HEADER_LEN + 1
	MAVLINK_NUM_CHECKSUM_BYTES    = 2
	MAVLINK_SIGNATURE_BLOCK_LEN   = 13
	MAVLINK_NUM_NON_PAYLOAD_BYTES = MAVLINK_NUM_HEADER_BYTES + MAVLINK_NUM_CHECKSUM_BYTES
	MAVLINK_MAX_PACKET_LEN        = MAVLINK_MAX_PAYLOAD_LEN + MAVLINK_NUM_NON_PAYLOAD_BYTES + MAVLINK_SIGNATURE_BLOCK_LEN
)

var PAYLOAD_SIZES_BY_MSG_ID = map[uint32]byte{
	0: PAYLOAD_HEARTBEAT_CAPACITY,
	1: PAYLOAD_SYS_STATUS_CAPACITY,
}

type Message struct {
	buffer  [MAVLINK_MAX_PACKET_LEN]byte
	Header  *Header  `json:"header"`
	Payload *Payload `json:"payload"`
	crc     *CRC
	len     int
}

func NewMessage() *Message {
	message := &Message{Header: nil, Payload: nil, crc: newCRC()}
	message.Header = NewHeader((*[MAVLINK_NUM_HEADER_BYTES]byte)(unsafe.Pointer(&message.buffer[0])))
	message.Payload = NewPayload((*[MAVLINK_MAX_PAYLOAD_LEN]byte)(unsafe.Pointer(&message.buffer[MAVLINK_NUM_HEADER_BYTES])), 0)
	return message
}

func NewMessageFrom(stx, payloadCapacity, seq, sysid, compId byte, msgId uint32) *Message {
	if payloadCapacity == 0 {
		if size, ok := PAYLOAD_SIZES_BY_MSG_ID[msgId]; ok {
			payloadCapacity = size
		} else {
			payloadCapacity = MAVLINK_MAX_PAYLOAD_LEN
		}
	}
	message := &Message{Header: nil, Payload: nil, crc: newCRC()}
	message.crc.Reset()
	ptr := (*[MAVLINK_NUM_HEADER_BYTES]byte)(unsafe.Pointer(&message.buffer[0]))
	message.SetHeader(NewHeaderWith(ptr, stx, payloadCapacity, seq, sysid, compId, msgId))

	message.Header.len = MAVLINK_NUM_HEADER_BYTES

	message.update(10, 0)

	message.Payload = NewPayload((*[MAVLINK_MAX_PAYLOAD_LEN]byte)(unsafe.Pointer(&message.buffer[MAVLINK_NUM_HEADER_BYTES])), payloadCapacity)
	return message
}

func (m *Message) GetHeader() *Header {
	return m.Header
}

func (m *Message) SetHeader(header *Header) {
	m.Header = header
}

func (m *Message) GetPayload() *Payload {
	return m.Payload
}

func (m *Message) SetPayload(payload *Payload) {
	m.Payload = payload
}

func (m *Message) GetCRC() uint16 {
	return m.crc.GetCRC()
}

func (m *Message) SetCRC(crc uint16) {
	m.crc.Raw = uint64(crc)
}

func (m *Message) Clear() {
	m.Header = nil
	m.Payload = nil
	m.crc = newCRC()
}

// returns true if the message is full and ready to be sent
func (m *Message) PushByte(value byte) (bool, error) {
	return m.PushBytes([]byte{value})
}

func (m *Message) PushUint16(value uint16) (bool, error) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	return m.PushBytes(buf[:])
}

func (m *Message) PushFloat32(value float32) (bool, error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(value))
	return m.PushBytes(buf[:])
}

func (m *Message) PushFloat64(value float64) (bool, error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(value))
	return m.PushBytes(buf[:])
}

func (m *Message) PushUint32(value uint32) (bool, error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	return m.PushBytes(buf[:])
}

func (m *Message) PushBytes(values []byte) (bool, error) {
	if m.len >= MAVLINK_MAX_PACKET_LEN {
		return false, errors.New("Message is full. Cannot push more bytes.")
	}
	valuesLen := len(values)
	copy(m.buffer[m.len:], values)
	return m.update(valuesLen, m.len), nil
}

func (m *Message) GetRawMessage() []byte {
	return m.buffer[:m.len]
}

func (m *Message) update(pushedSize int, pushedPosition int) bool {
	switch {
	case m.len == 0:
		if pushedPosition == 0 {
			m.crc.Reset()
		}
	case m.len < MAVLINK_NUM_HEADER_BYTES:
		m.Header.len += byte(pushedSize)
	case (m.len >= MAVLINK_NUM_HEADER_BYTES) && !m.Payload.IsFull():
		m.Payload.Len += byte(pushedSize)
	}

	if (m.len >= MAVLINK_NUM_HEADER_BYTES) && m.Payload.IsFull() {
		m.crc.Calculate(m.buffer[m.len : m.len+pushedSize])
		m.len += pushedSize
		binary.LittleEndian.PutUint16(m.buffer[m.len:m.len+2], m.crc.GetCRC())
		m.len += 2
		return true
	}

	m.crc.Calculate(m.buffer[m.len : m.len+pushedSize])
	m.len += pushedSize

	return false
}
