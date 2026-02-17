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
	MAVLINK_CRC_EXTRA_LEN         = 1
)

var PAYLOAD_SIZES_BY_MSG_ID = map[uint32]byte{
	0: PAYLOAD_HEARTBEAT_CAPACITY,
	1: PAYLOAD_SYS_STATUS_CAPACITY,
}

type Message struct {
	buffer         [MAVLINK_MAX_PACKET_LEN]byte
	Header         *Header  `json:"header"`
	Payload        *Payload `json:"payload"`
	crcExtraPushed bool
	crc            *CRC
	len            int
}

func NewMessage() *Message {
	message := &Message{Header: nil, Payload: nil, crc: newCRC()}
	message.crc.Reset()
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
	message.len = MAVLINK_NUM_HEADER_BYTES

	message.crc.Calculate(message.buffer[1:10])

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
	m.Header.len = 0
	m.Payload.Len = 0
	m.len = 0
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
	if values == nil {
		return false, errors.New("Cannot push nil byte slice.")
	}
	valuesLen := len(values)
	if valuesLen == 0 {
		return false, errors.New("No bytes to push.")
	}
	if (m.len >= MAVLINK_NUM_HEADER_BYTES) && (m.len+valuesLen > int(*m.Header.Len)+MAVLINK_NUM_HEADER_BYTES) {
		return false, errors.New("Not enough space in the message to push the given bytes.")
	}

	return m.update(values, valuesLen, m.len)
}

func (m *Message) GetRawMessage() []byte {
	return m.buffer[:m.len]
}

func (m *Message) update(values []byte, pushedSize int, pushedPosition int) (bool, error) {
	switch {
	case m.len == 0:
		m.crcExtraPushed = false
		m.updateHeader(values, pushedSize, pushedPosition)
	case m.len < MAVLINK_NUM_HEADER_BYTES:
		m.updateHeader(values, pushedSize, pushedPosition)
	case (m.len >= MAVLINK_NUM_HEADER_BYTES) && !m.Payload.IsFull():
		m.updatePayload(values, pushedSize, pushedPosition)
	case m.len+pushedSize == MAVLINK_NUM_HEADER_BYTES+int(m.Header.len)+1:
		m.updateCRCExtra(values[0])
	}
	return m.IsFull(), nil
}

func (m *Message) updateHeader(values []byte, pushedSize int, pushedPosition int) (bool, error) {
	if pushedSize == 0 {
		return false, nil
	}
	if pushedSize > MAVLINK_NUM_HEADER_BYTES {
		copy(m.buffer[:], values[:MAVLINK_NUM_HEADER_BYTES])
		return m.updatePayload(values[MAVLINK_NUM_HEADER_BYTES:pushedSize], pushedSize-MAVLINK_NUM_HEADER_BYTES, 0)
	}

	m.crc.Reset()

	return false, nil
}

func (m *Message) updatePayload(values []byte, pushedSize int, pushedPosition int) (bool, error) {
	lastIndex := m.len + pushedSize
	payloadSize := lastIndex - MAVLINK_NUM_HEADER_BYTES
	switch {
	case pushedSize == 0:
		return false, nil
	case payloadSize > int(*m.Header.Len)+1:
		return false, errors.New("Payload size exceeds maximum packet length")
	case payloadSize > int(*m.Header.Len)+1:
		return false, errors.New("Payload size exceeds maximum payload length")
	case payloadSize == int(*m.Header.Len)+1:
		crcExtra := byte(0)
		crcExtra = values[pushedSize-1]
		m.updateCRCExtra(crcExtra)
		values = values[:pushedSize-1]
		m.crc.Calculate(values[:pushedSize])
		copy(m.buffer[m.len:], values[:pushedSize])
		m.len += pushedSize
		return true, nil
	default:
		m.crc.Calculate(values[:pushedSize])
		copy(m.buffer[m.len:], values[:pushedSize])
		m.len += pushedSize
		return false, nil
	}
}

func (m *Message) updateCRCExtra(crcExtra byte) (bool, error) {
	if crcExtra == 0 {
		return false, nil
	}

	m.crc.Calculate([]byte{crcExtra})
	binary.LittleEndian.PutUint16(m.buffer[m.len:m.len+2], m.crc.GetCRC())
	m.len += 2
	m.crcExtraPushed = true
	return m.IsFull(), nil
}

func (m *Message) IsFull() bool {
	return m.Header.IsFull() && m.Payload.IsFull() && m.crcExtraPushed
}
