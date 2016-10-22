package message

import (
	"encoding/binary"
	"fmt"
)

type PacketRecord struct {
	Header MessageHeader
	RecordHeader RecordHeader
	DeviceID uint32
	EventID uint32
	EventSecond uint32
	PacketSecond uint32
	PacketMicrosecond uint32
	LinkType uint32
	PacketLength uint32
	PacketData []byte
}

func UnmarshalPacketRecord(header MessageHeader, recordHeader RecordHeader, body []byte) PacketRecord {
	var record PacketRecord

	header.copy( &record.Header )
	recordHeader.copy( &record.RecordHeader )

	record.DeviceID = binary.BigEndian.Uint32(body[:4])
	record.EventID = binary.BigEndian.Uint32(body[4:8])
	record.EventSecond = binary.BigEndian.Uint32(body[8:12])
	record.PacketSecond = binary.BigEndian.Uint32(body[12:16])
	record.PacketMicrosecond = binary.BigEndian.Uint32(body[16:20])
	record.LinkType = binary.BigEndian.Uint32(body[20:24])
	record.PacketLength = binary.BigEndian.Uint32(body[24:48])
	record.PacketData = body[28:]

	return record
}

func (record PacketRecord) String () string {
	return fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%X",
		record.Header.MessageType,
		record.RecordHeader.RecordType,
		record.RecordHeader.ServerTimestamp,
		record.DeviceID,
		record.EventID,
		record.EventSecond,
		record.PacketSecond,
		record.PacketMicrosecond,
		record.LinkType,
		record.PacketLength,
		record.PacketData,
	)
}