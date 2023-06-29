package slidewindow

import (
	"container/list"
	"sync"
	"time"
)

type Limiter struct {
	queue    *list.List
	interval int64
	rate     int
	mutex    sync.Mutex
	close    chan struct{}
}

func (l *Limiter) Allow() bool {
	now := time.Now().UnixNano()
	boundary := now - l.interval

	l.mutex.Lock()
	timestamp := l.queue.Front()

	// 把不在窗口的数据删除
	for timestamp != nil && timestamp.Value.(int64) < boundary {
		l.queue.Remove(timestamp)
		timestamp = l.queue.Front()
	}
	length := l.queue.Len()
	l.mutex.Unlock()

	if length >= l.rate {
		return false
	}

	l.queue.PushBack(now)
	return true
}

func (l *Limiter) Close() {
	once := sync.Once{}
	once.Do(func() {
		close(l.close)
	})
}

func NewLimiter(interval time.Duration, rate int) *Limiter {
	return &Limiter{
		queue:    list.New(),
		interval: interval.Nanoseconds(),
		rate:     rate,
		mutex:    sync.Mutex{},
		close:    make(chan struct{}),
	}
}
