package client

import (
	"context"
	"fmt"
)

type Req struct {
	ID int
}

type Resp struct {
	Msg string
}

type UserService struct {
	Get func(ctx context.Context, req *Req) (*Resp, error)
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
