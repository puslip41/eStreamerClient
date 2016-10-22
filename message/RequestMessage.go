package message

import (
	"time"
	"encoding/binary"
)

type RequestFlag uint32
const (
	BIT0 RequestFlag = 1 << iota
	BIT1
	BIT2
	BIT3
	BIT4
	BIT5
	BIT6
	BIT7
	BIT8
	BIT9
	BIT10
	BIT11
	BIT12
	BIT13
	BIT14
	BIT15
	BIT16
	BIT17
	BIT18
	BIT19
	BIT20
	BIT21
	BIT22
	BIT23
	BIT24
	BIT25
	BIT26
	BIT27
	BIT28
	BIT29
	BIT30
	BIT31
)

type MessageTypeCode uint16
const (
	NULL_MESSAGE MessageTypeCode = 0
	ERROR_MESSAGE = 1
	EVENT_STREAM_REQUEST = 2
	EVENT_DATA = 4
	HOST_DATA_REQUEST = 5
	SINGLE_HOST_DATA = 6
	MULTIPLE_HOST_DATA = 7
	STREAMING_REQUEST = 2049
	STREAMING_INFORMATION = 2051
	MESSAGE_BUNDLE = 4002
)

func GetMessageTypeCode(v uint16) MessageTypeCode {
	var code MessageTypeCode

	switch v {
	case 0:
		code = NULL_MESSAGE
	case 1:
	code = ERROR_MESSAGE
	case 2:
	code = EVENT_STREAM_REQUEST
	case 4:
		code = EVENT_DATA
	case 5:
		code = HOST_DATA_REQUEST
	case 6:
		code = SINGLE_HOST_DATA
	case 7:
		code = MULTIPLE_HOST_DATA
	case 2049:
		code = STREAMING_REQUEST
	case 2051:
		code = STREAMING_INFORMATION
	case 4002:
		code = MESSAGE_BUNDLE
		default:
	}

	return code
}

type RequestMessage struct {
	Header MessageHeader
	InitialTimestamp time.Time
	RequestFlags RequestFlag
}

func (message *RequestMessage) Marshal() []byte {
	bytes := make([]byte, 16)

	copy(bytes, message.Header.Marshal())

	uint64Bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(uint64Bytes, uint64(message.InitialTimestamp.Unix()))
	copy(bytes[8:12], uint64Bytes[4:8])
	binary.BigEndian.PutUint32(bytes[12:16], uint32(message.RequestFlags))

	return bytes
}

func UnmarshalHeader(bytes []byte) MessageHeader {
	return MessageHeader{
		HeaderVersion:binary.BigEndian.Uint16(bytes[:2]),
		MessageType:GetMessageTypeCode(binary.BigEndian.Uint16(bytes[2:4])),
		MessageLength:binary.BigEndian.Uint32(bytes[4:8]),
	}
}

func (v *RawMessage) String() string {
	return ""
}

type RawMessage struct {
	Header MessageHeader
	Content []byte
}
