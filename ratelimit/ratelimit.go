package ratelimit

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Limiter interface {
	Allow() bool
	Close()
}

func NewServerLimiter(limiter Limiter) grpc.ServerOption {
	interceptor := BuildServerInterceptor(limiter)
	return grpc.UnaryInterceptor(interceptor)
}

func NewClientLimiter(limiter Limiter) grpc.DialOption {
	interceptor := BuildClientInterceptor(limiter)
	return grpc.WithUnaryInterceptor(interceptor)
}

func BuildServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !limiter.Allow() {
			return nil, errors.New("rate-limit")
		}
		resp, err = handler(ctx, req)
		return
	}
}

func BuildClientInterceptor(limiter Limiter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !limiter.Allow() {
			return errors.New("rate-limit")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
