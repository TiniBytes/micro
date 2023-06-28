package ratelimit

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Limiter interface {
	Allow() bool
	Close()
}

func NewServerLimiter(limiter Limiter) grpc.UnaryServerInterceptor {
	return BuildServerInterceptor(limiter)
}

func BuildServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !limiter.Allow() {
			return nil, errors.New("rate-limit")
		}
		return handler(ctx, req), nil
	}
}
