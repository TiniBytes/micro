package client

import (
	"context"
	"fmt"
	"log"
	"micro/demo/proto"
	"testing"
	"time"
)

type Req struct {
	ID int
}

type Resp struct {
	Msg string
}

type UserService struct {
	Get          func(ctx context.Context, req *Req) (*Resp, error)
	GetByIDProto func(ctx context.Context, req *proto.GetByIDReq) (*proto.GetByIDResp, error)
	GetTimeout   func(ctx context.Context, req *proto.GetByIDReq) (*proto.GetByIDResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}

func (u *UserServiceServer) Get(ctx context.Context, req *Req) (*Resp, error) {
	fmt.Println(req)
	return &Resp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIDProto(ctx context.Context, req *proto.GetByIDReq) (*proto.GetByIDResp, error) {
	log.Println(req)
	return &proto.GetByIDResp{
		User: &proto.User{
			Name: u.Msg,
		},
	}, u.Err
}

type UserServiceTimeout struct {
	t     *testing.T
	Err   error
	Msg   string
	sleep time.Duration
}

func (u *UserServiceTimeout) GetTimeout(ctx context.Context, req *proto.GetByIDReq) (*proto.GetByIDResp, error) {
	if _, ok := ctx.Deadline(); ok {
		u.t.Fatal("未设置超时")
	}
	time.Sleep(u.sleep)
	return &proto.GetByIDResp{
		User: &proto.User{
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceTimeout) Name() string {
	return "user-service"
}
