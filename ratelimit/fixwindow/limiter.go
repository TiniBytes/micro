package fixwindow

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type Limiter struct {
	timestamp int64
	interval  int64
	rate      int64
	cnt       int64
}

func NewLimiter(interval time.Duration, rate int64) *Limiter {
	return &Limiter{
		timestamp: time.Now().UnixNano(),
		interval:  interval.Nanoseconds(),
		rate:      rate,
		cnt:       0,
	}
}

func (f *Limiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// window reset
		current := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)
		if current > f.timestamp+f.interval {
			// new window, reset windows
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, current) {
				atomic.CompareAndSwapInt64(&f.cnt, cnt, 0)
			}
		}

		cnt = atomic.AddInt64(&f.cnt, 1)
		if cnt > f.rate {
			err = errors.New("rate-limit")
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}
