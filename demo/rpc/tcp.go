package rpc

import (
	"encoding/binary"
	"net"
)

const numOfLengthBytes = 8

// ReadMsg 读取消息
func ReadMsg(conn net.Conn) ([]byte, error) {
	byteLen := make([]byte, numOfLengthBytes)
	_, err := conn.Read(byteLen) // 读取长度
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint64(byteLen)
	data := make([]byte, length)
	_, err = conn.Read(data) // 根据长度读数据
	return data, err
}

// EncodeMsg 消息编码
func EncodeMsg(date []byte) []byte {
	rspLen := len(date)
	res := make([]byte, numOfLengthBytes+rspLen)

	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(rspLen))
	copy(res[numOfLengthBytes:], date)
	return res
}
