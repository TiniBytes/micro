package ratelimit

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})

	producer := time.NewTicker(interval)
	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C:
				// Put in the token
				ch <- struct{}{}
			case <-closeCh:
				return
			default:

			}
		}
	}()

	return &TokenBucketLimiter{
		tokens: ch,
		close:  closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		select {
		case <-t.close:
			// No current limit on
			resp, err = handler(ctx, req)
		case <-t.tokens:
			// Get the token
			resp, err = handler(ctx, req)
		case <-ctx.Done():
			// Trigger current limit
			err = ctx.Err()
			return
		}
		return
	}
}

func (t *TokenBucketLimiter) Close() error {
	once := sync.Once{}
	once.Do(func() {
		close(t.close)
	})

	return nil
}
