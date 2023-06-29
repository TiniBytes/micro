package middleware

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type HandleFunc func(ctx ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor

// Middleware 函数式责任链模式
type Middleware func()

func BuildServerInterceptor(middleware Middleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		middleware()
		resp, err = handler(ctx, req)
		return
	}
}
