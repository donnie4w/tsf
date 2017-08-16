package utils

import (
	"bytes"
	"encoding/binary"
)

func Int64ToBytes(i int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return bytesBuffer.Bytes()
}

func ByteToInt64(b []byte) (i int64) {
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, &i)
	return
}

func IntToBytes(i int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) (i int32) {
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, &i)
	return
}

func ShortToBytes(i int16) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return bytesBuffer.Bytes()
}

func BytesToShort(b []byte) (i int16) {
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, &i)
	return
}
