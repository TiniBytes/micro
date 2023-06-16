package protocol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeResponse(t *testing.T) {

	tests := []struct {
		name string
		resp *Response
	}{
		// TODO: Add test cases.
		{
			name: "case1_normal",
			resp: &Response{
				//HeadLength: 11,
				//BodyLength: 12,
				MessageID:  13,
				Version:    14,
				Compress:   15,
				Serializer: 16,
				Error:      []byte("error"),
				Data:       []byte("hello,world"),
			},
		},
		{
			name: "case2_no_error",
			resp: &Response{
				//HeadLength: 11,
				//BodyLength: 12,
				MessageID:  13,
				Version:    14,
				Compress:   15,
				Serializer: 16,
				//Error:      []byte("error"),
				Data: []byte("hello,world"),
			},
		},
		{
			name: "case3_no_data",
			resp: &Response{
				//HeadLength: 11,
				//BodyLength: 12,
				MessageID:  13,
				Version:    14,
				Compress:   15,
				Serializer: 16,
				Error:      []byte("error"),
				//Data: []byte("hello,world"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalculateHeaderLength()
			tc.resp.CalculateBodyLength()

			data := EncodeResponse(tc.resp)
			resp := DecodeResponse(data)
			fmt.Println(resp)
			assert.Equal(t, tc.resp, resp)
		})
	}
}
