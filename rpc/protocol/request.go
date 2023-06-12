package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	splitter     = '\n'
	pairSplitter = '\r'

	HeadLengthBytes      = 4
	BodyLengthBytes      = 4
	MessageIDLengthBytes = 4
)

type Request struct {
	HeadLength  uint32
	BodyLength  uint32
	MessageID   uint32
	Version     uint8
	Compress    uint8
	Serializer  uint8
	ServiceName string
	MethodName  string
	Meta        map[string]string
	Data        []byte
}

func EncodeRequest(req *Request) []byte {
	data := make([]byte, req.HeadLength+req.BodyLength)
	cur := data

	binary.BigEndian.PutUint32(cur[:HeadLengthBytes], req.HeadLength)
	cur = cur[HeadLengthBytes:]

	binary.BigEndian.PutUint32(cur[:BodyLengthBytes], req.BodyLength)
	cur = cur[BodyLengthBytes:]

	binary.BigEndian.PutUint32(cur[:MessageIDLengthBytes], req.MessageID)
	cur = cur[MessageIDLengthBytes:]

	cur[0] = req.Version
	cur = cur[1:]

	cur[0] = req.Compress
	cur = cur[1:]

	cur[0] = req.Serializer
	cur = cur[1:]

	copy(cur, req.ServiceName)
	cur[len(req.ServiceName)] = splitter
	cur = cur[len(req.ServiceName)+1:]

	copy(cur, req.MethodName)
	cur[len(req.MethodName)] = splitter
	cur = cur[len(req.MethodName)+1:]

	return data
}

func DecodeRequest(data []byte) *Request {
	request := &Request{}

	request.HeadLength = binary.BigEndian.Uint32(data[:HeadLengthBytes])
	data = data[HeadLengthBytes:]

	request.BodyLength = binary.BigEndian.Uint32(data[:BodyLengthBytes])
	data = data[BodyLengthBytes:]

	request.MessageID = binary.BigEndian.Uint32(data[:MessageIDLengthBytes])
	data = data[MessageIDLengthBytes:]

	request.Version = data[0]
	data = data[1:]

	request.Compress = data[0]
	data = data[1:]

	request.Serializer = data[0]
	data = data[1:]

	index := bytes.IndexByte(data, splitter)
	request.ServiceName = string(data[:index])
	data = data[index+1:]

	index = bytes.IndexByte(data, splitter)
	request.MethodName = string(data[:index])
	data = data[index+1:]

	return request
}
