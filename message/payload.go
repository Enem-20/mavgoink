package message

import (
	"errors"
	"strconv"
)

const (
	PAYLOAD_HEARTBEAT_CAPACITY                    = 9
	PAYLOAD_SYS_STATUS_CAPACITY                   = 43
	PAYLOAD_SYSTEM_TIME_CAPACITY                  = 12
	PAYLOAD_PING_CAPACITY                         = 14
	PAYLOAD_CHANGE_OPERATOR_CONTROL_CAPACITY      = 29
	PAYLOAD_CHANGE_OPERATOR_CONTROL_ACK_CAPACITY  = 33
	PAYLOAD_AUTH_KEY_CAPACITY                     = 32
	PAYLOAD_LINK_NODE_STATUS_CAPACITY             = 36
	PAYLOAD_SET_MODE_CAPACITY                     = 6
	PAYLOAD_PARAM_REQUEST_READ_CAPACITY           = 20
	PAYLOAD_PARAM_REQUEST_LIST_CAPACITY           = 2
	PAYLOAD_PARAM_VALUE_CAPACITY                  = 25
	PAYLOAD_PARAM_SET_CAPACITY                    = 23
	PAYLOAD_GPS_RAW_INT_CAPACITY                  = 54
	PAYLOAD_GPS_STATUS_CAPACITY                   = 101
	PAYLOAD_SCALED_IMU_CAPACITY                   = 22
	PAYLOAD_RAW_IMU_CAPACITY                      = 24
	PAYLOAD_RAW_PRESSURE_CAPACITY                 = 16
	PAYLOAD_SCALED_PRESSURE_CAPACITY              = 22
	PAYLOAD_ATTITUDE_CAPACITY                     = 28
	PAYLOAD_ATTITUDE_QUATERNION_CAPACITY          = 48
	PAYLOAD_LOCAL_POSITION_NED_CAPACITY           = 32
	PAYLOAD_GLOBAL_POSITION_INT_CAPACITY          = 28
	PAYLOAD_RC_CHANNELS_SCALED_CAPACITY           = 23
	PAYLOAD_RC_CHANNELS_RAW_CAPACITY              = 22
	PAYLOAD_SERVO_OUTPUT_RAW_CAPACITY             = 37
	PAYLOAD_MISSION_REQUEST_PARTIAL_LIST_CAPACITY = 7
	PAYLOAD_MISSION_WRITE_PARTIAL_LIST_CAPACITY   = 7
	PAYLOAD_MISSION_ITEM_CAPACITY                 = 38
	PAYLOAD_MISSION_REQUEST_CAPACITY              = 5
	PAYLOAD_MISSION_SET_CURRENT_CAPACITY          = 4
	PAYLOAD_MISSION_CURRENT_CAPACITY              = 18
	PAYLOAD_MISSION_REQUEST_LIST_CAPACITY         = 3
	PAYLOAD_MISSION_COUNT_CAPACITY                = 9
	PAYLOAD_MISSION_CLEAR_ALL_CAPACITY            = 3
	PAYLOAD_MISSION_ITEM_REACHED_CAPACITY         = 2
	PAYLOAD_MISSION_ACK_CAPACITY                  = 8
	PAYLOAD_SET_GPS_GLOBAL_ORIGIN_CAPACITY        = 21
	PAYLOAD_GPS_GLOBAL_ORIGIN_CAPACITY            = 20
	PAYLOAD_PARAM_MAP_RC_CAPACITY                 = 37
	PAYLOAD_MISSION_REQUEST_INT_CAPACITY          = 5
	PAYLOAD_SET_ALLOWED_AREA_CAPACITY             = 27
	PAYLOAD_ALLOWED_AREA_CAPACITY                 = 25
	PAYLOAD_ATTITUDE_QUATERNION_COV_CAPACITY      = 84
	PAYLOAD_NAV_CONTROLLER_OUTPUT_CAPACITY        = 26
	PAYLOAD_GLOBAL_POSITION_INT_COV_CAPACITY      = 36
)

type Payload struct {
	Data     *[MAVLINK_MAX_PAYLOAD_LEN]byte `json:"data"`
	Len      byte                           `json:"len"`
	Capacity byte                           `json:"capacity"`
}

func NewPayload(data *[MAVLINK_MAX_PAYLOAD_LEN]byte, capacity byte) *Payload {
	return &Payload{Data: data, Len: 0, Capacity: capacity}
}

func (p *Payload) GetRawPayload() []byte {
	return p.Data[:p.Len]
}

func (p *Payload) SetData(data [MAVLINK_MAX_PAYLOAD_LEN]byte) {
	copy(p.Data[:], data[:])
	p.Len = byte(len(data))
}

func (p *Payload) GetLength() byte {
	return p.Len
}

func (p *Payload) GetByte(index byte) byte {
	if index >= p.Len {
		return 0
	}
	return p.Data[index]
}

func (p *Payload) SetByte(index byte, value byte) error {
	if index >= p.Len {
		return errors.New("Out of bound. Illegal access to index " + strconv.Itoa(int(index)) + " in bounds [0, " + strconv.Itoa(int(p.Len)) + ")")
	}
	p.Data[index] = value
	return nil
}

func (p *Payload) AppendByte(value byte) {
	p.Data[p.Len] = value
	p.Len++
}

func (p *Payload) AppendBytes(values []byte) {
	for _, v := range values {
		p.AppendByte(v)
	}
}

func (p *Payload) Clear() {
	p.Len = 0
}

func (p *Payload) IsFull() bool {
	return p.Len >= p.Capacity
}
