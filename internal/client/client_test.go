package client

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"micro/demo/proto"
	"micro/internal/server"
	proto2 "micro/rpc/serialize/proto"
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

	//time.Sleep(3 * time.Second)

	// 客户端
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

func TestInitServiceProto(t *testing.T) {
	svc := server.InitServer("tcp", ":8080")
	service := &UserServiceServer{}
	svc.RegisterService(service)
	svc.RegisterSerializer(&proto2.Serializer{})
	go func() {
		err := svc.Start()
		t.Log(err)
	}()
	time.Sleep(3 * time.Second)

	// 客户端
	usClient := &UserService{}
	err := InitClientProxy(":8080", usClient, ClientWithSerializer(&proto2.Serializer{}))
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
		{
			name: "case4_timeout",
			mock: func() {
				service.Err = nil
				service.Msg = "hello,world"
			},
			wantResp: &Resp{
				Msg: "hello,world",
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
			resp, er := usClient.GetByIDProto(ctx, &proto.GetByIDReq{
				Id: 1231313,
			})

			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}

}
