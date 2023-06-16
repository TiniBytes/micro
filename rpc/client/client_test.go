package client

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"micro/rpc/server"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	svc := server.InitServer("tcp", ":8080")
	service := &UserServiceServer{}
	svc.RegisterService(service)
	go func() {
		err := svc.Start()
		t.Log(err)
	}()
	time.Sleep(3 * time.Second)

	usClient := &UserService{}
	err := InitClientProxy(":8080", usClient)
	if err != nil {
		return
	}

	tests := []struct {
		name     string
		mock     func()
		wantErr  error
		wantResp *Resp
	}{
		{
			name: "case1_no_err",
			mock: func() {
				service.Err = nil
				service.Msg = "hello,world"
			},
			wantResp: &Resp{
				Msg: "hello,world",
			},
		},
		{
			name: "case2_no_msg",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = ""
			},
			wantResp: &Resp{},
			wantErr:  errors.New("mock error"),
		},
		{
			name: "case3_err_msg",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello,world"
			},
			wantResp: &Resp{
				Msg: "hello,world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.Get(context.Background(), &Req{ID: 13})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}

	// 发起调用
	//resp, err := usClient.Get(context.Background(), &Req{ID: 13})
	//if err != nil {
	//	return
	//}
	//
	//assert.Equal(t, &Resp{
	//	Msg: "hello,world",
	//}, resp)
}