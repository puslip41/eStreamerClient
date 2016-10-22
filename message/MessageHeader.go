package message

import "encoding/binary"

type MessageHeader struct {
	HeaderVersion uint16
	MessageType MessageTypeCode
	MessageLength uint32
}

func (header MessageHeader) copy (v *MessageHeader) {
	if v != nil {
		v.MessageType = header.MessageType
		v.MessageLength = header.MessageLength
		v.HeaderVersion = header.HeaderVersion
	}
}

func (header *MessageHeader) Marshal() []byte {
	bytes := make([]byte, 8)

	binary.BigEndian.PutUint16(bytes[:2], uint16(header.HeaderVersion))
	binary.BigEndian.PutUint16(bytes[2:4], uint16(header.MessageType))
	binary.BigEndian.PutUint32(bytes[4:8], uint32(header.MessageLength))

	return bytes
}
