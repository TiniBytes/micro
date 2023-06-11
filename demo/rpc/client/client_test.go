package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"micro/demo/rpc/server"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	svc := server.InitServer("tcp", ":8080")
	svc.RegisterService(&UserServiceServer{})
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

	// 发起调用
	resp, err := usClient.Get(context.Background(), &Req{ID: 13})
	if err != nil {
		return
	}

	assert.Equal(t, &Resp{
		Msg: "hello",
	}, resp)
}

type UserServiceServer struct{}

func (u *UserServiceServer) Name() string {
	return "user-service"
}

func (u *UserServiceServer) Get(ctx context.Context, req *Req) (*Resp, error) {
	log.Println(req)
	return &Resp{
		Msg: "hello",
	}, nil
}
