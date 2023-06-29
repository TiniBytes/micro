package tokenbucket

import (
	"sync"
	"time"
)

type Limiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func (l *Limiter) Allow() bool {
	select {
	case <-l.close:
		// No current limit on
		return true
	case <-l.tokens:
		// Get the token
		return true
	default:
		return false
	}
}

func (l *Limiter) Close() {
	once := sync.Once{}
	once.Do(func() {
		close(l.close)
	})
}

func NewLimiter(capacity int, interval time.Duration) *Limiter {
	ch := make(chan struct{}, capacity)
	ch <- struct{}{}
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
