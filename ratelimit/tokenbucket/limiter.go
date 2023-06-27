package tokenbucket

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Limiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewLimiter(capacity int, interval time.Duration) *Limiter {
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

	return &Limiter{
		tokens: ch,
		close:  closeCh,
	}
}

func (t *Limiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
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

func (t *Limiter) Close() error {
	once := sync.Once{}
	once.Do(func() {
		close(t.close)
	})

	return nil
}
