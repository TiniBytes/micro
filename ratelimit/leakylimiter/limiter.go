package leakylimiter

import (
	"time"
)

type Limiter struct {
	producer *time.Ticker
}

func (l *Limiter) Allow() bool {
	select {
	case <-l.producer.C:
		return true
	default:
		return false
	}
}

func (l *Limiter) Close() {
	l.producer.Stop()
}

func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{
		producer: time.NewTicker(interval),
	}
}
