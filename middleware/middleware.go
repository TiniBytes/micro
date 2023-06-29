package middleware

import (
	"context"
	"google.golang.org/grpc"
)

type Handler func(ctx context.Context, info interface{}) (reply interface{}, err error)

type Middleware func(handler Handler) Handler

func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}

func BuildServerInterceptor(m []Middleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		h := func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			return handler(ctx, req)
		}

		if len(m) > 0 {
			h = Chain(m...)(h)
		}
		reply, err := h(ctx, req)
		return reply, err
	}
}
