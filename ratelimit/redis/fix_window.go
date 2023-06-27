package redis

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixWindow string

type FixWindowLimiter struct {
	client   redis.Cmdable
	interval time.Duration
	rate     int
	service  string
}

func NewFixWindowLimiter(client redis.Cmdable, interval time.Duration, rate int, service string) *FixWindowLimiter {
	return &FixWindowLimiter{
		client:   client,
		interval: interval,
		rate:     rate,
		service:  service,
	}
}

func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 不同限流粒度
		allow, err := f.allow(ctx)
		if err != nil {
			return nil, err
		}
		if !allow {
			err = errors.New("rate-limit")
			return
		}

		resp, err = handler(ctx, req)
		return
	}
}

func (f *FixWindowLimiter) allow(ctx context.Context) (bool, error) {
	return f.client.Eval(ctx, luaFixWindow, []string{f.service}).Bool()
}
