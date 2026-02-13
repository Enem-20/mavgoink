package message

import "unsafe"

const (
	HEADER_SIZE = 10
)

type Header struct {
	Stx      *byte    `json:"stx"`
	Len      *byte    `json:"len"`
	IncFlags *byte    `json:"incFlags"`
	CmpFlags *byte    `json:"cmpFlags"`
	Seq      *byte    `json:"seq"`
	Sysid    *byte    `json:"sysid"`
	CompID   *byte    `json:"compid"`
	MsgidArr *[3]byte `json:"msgidArr"`
	buffer   *[HEADER_SIZE]byte
	len      byte
}

func NewHeader(buffer *[HEADER_SIZE]byte) *Header {
	header := &Header{
		buffer: buffer,
		len:    0,
	}
	header.Stx = &header.buffer[0]
	header.Len = &header.buffer[1]
	header.IncFlags = &header.buffer[2]
	header.CmpFlags = &header.buffer[3]
	header.Seq = &header.buffer[4]
	header.Sysid = &header.buffer[5]
	header.CompID = &header.buffer[6]
	header.MsgidArr = (*[3]byte)(unsafe.Pointer(&header.buffer[7]))
	return header
}

func NewHeaderWith(buffer *[HEADER_SIZE]byte, stx, len, seq, sysid, compid byte, msgid uint32) *Header {
	header := &Header{
		buffer: buffer,
	}
	header.Stx = &header.buffer[0]
	header.Len = &header.buffer[1]
	header.IncFlags = &header.buffer[2]
	header.CmpFlags = &header.buffer[3]
	header.Seq = &header.buffer[4]
	header.Sysid = &header.buffer[5]
	header.CompID = &header.buffer[6]
	header.MsgidArr = (*[3]byte)(unsafe.Pointer(&header.buffer[7]))
	header.fill([HEADER_SIZE]byte{stx, len, 0, 0, seq, sysid, compid, byte(msgid & 0xFF), byte((msgid >> 8) & 0xFF), byte((msgid >> 16) & 0xFF)})
	header.len = HEADER_SIZE
	return header
}

func (h *Header) GetSTX() byte {
	return *h.Stx
}

func (h *Header) SetSTX(stx byte) {
	h.buffer[0] = stx
}

func (h *Header) GetLen() byte {
	return *h.Len
}

func (h *Header) SetLen(len byte) {
	h.buffer[1] = len
}

func (h *Header) GetSeq() byte {
	return *h.Seq
}

func (h *Header) SetSeq(seq byte) {
	h.buffer[2] = seq
}

func (h *Header) GetSysID() byte {
	return *h.Sysid
}

func (h *Header) SetSysID(sysid byte) {
	h.buffer[3] = sysid
}

func (h *Header) GetCompID() byte {
	return *h.CompID
}

func (h *Header) SetCompID(compid byte) {
	h.buffer[4] = compid
}

func (h *Header) GetMsgID() uint32 {
	return uint32((*h.MsgidArr)[0]) | uint32((*h.MsgidArr)[1])<<8 | uint32((*h.MsgidArr)[2])<<16
}

func (h *Header) SetMsgID(msgid uint32) {
	h.buffer[5] = byte(msgid & 0xFF)
	h.buffer[6] = byte((msgid >> 8) & 0xFF)
	h.buffer[7] = byte((msgid >> 16) & 0xFF)
}

func (h *Header) GetRawHeader() []byte {
	return h.buffer[:HEADER_SIZE]
}

func CheckSTX(stx byte) bool {
	return stx == 0xFE || stx == 0xFD
}

func STXVersion(stx byte) byte {
	switch stx {
	case 0xFE:
		return 1
	case 0xFD:
		return 2
	default:
		return 0
	}
}

func (h *Header) pushByte(b byte) {
	if h.len < HEADER_SIZE {
		h.buffer[h.len] = b
		h.len++
	}
}

func (h *Header) IsFull() bool {
	return h.len >= HEADER_SIZE
}

func (h *Header) fill(data [HEADER_SIZE]byte) {
	copy(h.buffer[:], data[:])
	h.len = HEADER_SIZE
}
