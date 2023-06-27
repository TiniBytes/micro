package slidewindow

import (
	"container/list"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Limiter struct {
	queue    *list.List
	interval int64
	rate     int
	mutex    sync.Mutex
}

func NewLimiter(interval time.Duration, rate int) *Limiter {
	return &Limiter{
		queue:    list.New(),
		interval: interval.Nanoseconds(),
		rate:     rate,
		mutex:    sync.Mutex{},
	}
}

func (s *Limiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now().UnixNano()
		boundary := now - s.interval

		s.mutex.Lock()
		timestamp := s.queue.Front()

		// 把不在窗口的数据删除
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			s.queue.Remove(timestamp)
			timestamp = s.queue.Front()
		}
		length := s.queue.Len()
		s.mutex.Unlock()

		if length >= s.rate {
			err = errors.New("rate-limit")
			return
		}

		s.queue.PushBack(now)
		resp, err = handler(ctx, req)
		return
	}
}
