package message

import "encoding/binary"

type RecordHeader struct {
	RecordType uint32
	RecordLength uint32
	ServerTimestamp uint32
	Reserved uint32
}

func UnmarshalRecordHeader(bytes []byte) RecordHeader {
	return RecordHeader{
		RecordType:binary.BigEndian.Uint32(bytes[:4]),
		RecordLength:binary.BigEndian.Uint32(bytes[4:8]),
		ServerTimestamp:binary.BigEndian.Uint32(bytes[8:12]),
		Reserved:binary.BigEndian.Uint32(bytes[12:16]),
	}
}

func (header RecordHeader) copy (v *RecordHeader) {
	if v != nil {
		v.RecordType = header.RecordType
		v.RecordLength = header.RecordLength
		v.ServerTimestamp = header.ServerTimestamp
		v.Reserved = header.Reserved
	}
}
