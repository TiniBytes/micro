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

	for key, value := range req.Meta {
		copy(cur, key)
		cur[len(key)] = pairSplitter
		cur = cur[len(key)+1:]

		copy(cur, value)
		cur[len(value)] = splitter
		cur = cur[len(value)+1:]
	}

	copy(cur, req.Data)

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
	header := data[1 : request.HeadLength-14]

	index := bytes.IndexByte(header, splitter)
	request.ServiceName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, splitter)
	request.MethodName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, splitter)
	var meta map[string]string
	if index != -1 {
		meta = make(map[string]string, 0)
	}
	for index != -1 {
		pair := header[:index]
		pairIndex := bytes.IndexByte(pair, pairSplitter)
		key := string(pair[:pairIndex])
		value := string(pair[pairIndex+1:])
		meta[key] = value

		header = header[index+1:]
		index = bytes.IndexByte(header, splitter)
	}
	request.Meta = meta

	if request.BodyLength != 0 {
		request.Data = data[request.HeadLength-14:]
	}

	return request
}

func (r *Request) CalculateHeaderLength() {
	headLen := 15 + len(r.ServiceName) + len(r.MethodName) + 2
	for key, value := range r.Meta {
		headLen += len(key)
		headLen++
		headLen += len(value)
		headLen++
	}
	r.HeadLength = uint32(headLen)
}

func (r *Request) CalculateBodyLength() {
	r.BodyLength = uint32(len(r.Data))
}
