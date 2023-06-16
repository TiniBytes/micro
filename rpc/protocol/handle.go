package protocol

import (
	"encoding/binary"
	"net"
)

const (
	numOfLengthBytes = 8
)

// ReadMsg 读取消息
func ReadMsg(conn net.Conn) ([]byte, error) {
	// 读协议头和协议体长度
	byteLen := make([]byte, HeadLengthBytes+BodyLengthBytes)
	_, err := conn.Read(byteLen)
	if err != nil {
		return nil, err
	}
	headLen := binary.BigEndian.Uint32(byteLen[:HeadLengthBytes])
	bodyLen := binary.BigEndian.Uint32(byteLen[HeadLengthBytes:])

	// 读数据
	data := make([]byte, headLen+bodyLen)
	copy(data[:HeadLengthBytes+BodyLengthBytes], byteLen)
	_, err = conn.Read(data[HeadLengthBytes+BodyLengthBytes:])
	return data, err
}
