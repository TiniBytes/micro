package ratelimit

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	producer *time.Ticker
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		producer: time.NewTicker(interval),
	}
}

func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		select {
		case <-l.producer.C:
			resp, err = handler(ctx, req)
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
		return
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.producer.Stop()
	return nil
}
