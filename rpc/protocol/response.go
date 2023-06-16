package protocol

import (
	"encoding/binary"
)

type Response struct {
	HeadLength uint32
	BodyLength uint32
	MessageID  uint32
	Version    uint8
	Compress   uint8
	Serializer uint8
	Error      []byte
	Data       []byte
}

func EncodeResponse(resp *Response) []byte {
	data := make([]byte, resp.HeadLength+resp.BodyLength)
	cur := data

	binary.BigEndian.PutUint32(cur[:HeadLengthBytes], resp.HeadLength)
	cur = cur[HeadLengthBytes:]

	binary.BigEndian.PutUint32(cur[:BodyLengthBytes], resp.BodyLength)
	cur = cur[BodyLengthBytes:]

	binary.BigEndian.PutUint32(cur[:MessageIDLengthBytes], resp.MessageID)
	cur = cur[MessageIDLengthBytes:]

	cur[0] = resp.Version
	cur = cur[1:]

	cur[0] = resp.Compress
	cur = cur[1:]

	cur[0] = resp.Serializer
	cur = cur[1:]

	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]

	copy(cur, resp.Data)

	return data
}

func DecodeResponse(data []byte) *Response {
	response := &Response{}

	response.HeadLength = binary.BigEndian.Uint32(data[:HeadLengthBytes])
	data = data[HeadLengthBytes:]

	response.BodyLength = binary.BigEndian.Uint32(data[:BodyLengthBytes])
	data = data[BodyLengthBytes:]

	response.MessageID = binary.BigEndian.Uint32(data[:MessageIDLengthBytes])
	data = data[MessageIDLengthBytes:]

	response.Version = data[0]
	data = data[1:]

	response.Compress = data[0]
	data = data[1:]

	response.Serializer = data[0]
	data = data[1:]

	if response.HeadLength > 15 {
		response.Error = data[:response.HeadLength-15]
		data = data[response.HeadLength-15:]
	}

	if response.BodyLength != 0 {
		response.Data = data
	}

	return response
}

func (r *Response) CalculateHeaderLength() {
	r.HeadLength = 15 + uint32(len(r.Error))
}

func (r *Response) CalculateBodyLength() {
	r.BodyLength = uint32(len(r.Data))
}
