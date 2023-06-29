package fixwindow

import (
	"sync"
	"sync/atomic"
	"time"
)

type Limiter struct {
	timestamp int64
	interval  int64
	rate      int64
	cnt       int64
	close     chan struct{}
}

func (l *Limiter) Allow() bool {
	// window reset
	current := time.Now().UnixNano()
	timestamp := atomic.LoadInt64(&l.timestamp)
	cnt := atomic.LoadInt64(&l.cnt)
	if current > l.timestamp+l.interval {
		// new window, reset windows
		if atomic.CompareAndSwapInt64(&l.timestamp, timestamp, current) {
			atomic.CompareAndSwapInt64(&l.cnt, cnt, 0)
		}
	}

	cnt = atomic.AddInt64(&l.cnt, 1)
	if cnt > l.rate {
		return false
	}

	return true
}

func (l *Limiter) Close() {
	once := sync.Once{}
	once.Do(func() {
		close(l.close)
	})
}

func NewLimiter(interval time.Duration, rate int64) *Limiter {
	return &Limiter{
		timestamp: time.Now().UnixNano(),
		interval:  interval.Nanoseconds(),
		rate:      rate,
		cnt:       0,
		close:     make(chan struct{}),
	}
}
