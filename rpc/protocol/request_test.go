package protocol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		req  *Request
	}{
		// TODO 测试用例
		{
			name: "case1_normal",
			req: &Request{
				//HeadLength:  24,
				//BodyLength:  48,
				MessageID:   11,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("hello, word"),
			},
		},
		{
			name: "case2_no mata with data",
			req: &Request{
				//HeadLength:  24,
				//BodyLength:  48,
				MessageID:   11,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Data:        []byte("hello, word"),
			},
		},
		{
			name: "case3_no data with meta",
			req: &Request{
				//HeadLength:  24,
				//BodyLength:  48,
				MessageID:   11,
				Version:     12,
				Compress:    13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.calculateHeaderLength()
			tc.req.calculateBodyLength()

			data := EncodeRequest(tc.req)
			req := DecodeRequest(data)
			fmt.Println(req)
			assert.Equal(t, tc.req, req)
		})
	}
}

func (r *Request) calculateHeaderLength() {
	headLen := 15 + len(r.ServiceName) + len(r.MethodName) + 2
	for key, value := range r.Meta {
		headLen += len(key)
		headLen++
		headLen += len(value)
		headLen++
	}
	r.HeadLength = uint32(headLen)
}

func (r *Request) calculateBodyLength() {
	r.BodyLength = uint32(len(r.Data))
}
