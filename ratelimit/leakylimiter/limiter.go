package leakylimiter

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type Limiter struct {
	producer *time.Ticker
}

func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{
		producer: time.NewTicker(interval),
	}
}

func (l *Limiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
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

func (l *Limiter) Close() error {
	l.producer.Stop()
	return nil
}
