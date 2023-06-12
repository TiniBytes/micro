package protocol

import "context"

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}
