package client

import "context"

// test ---------------------------------
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
