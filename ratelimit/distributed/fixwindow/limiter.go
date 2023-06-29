package fixwindow

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixWindow string

type Limiter struct {
	client   redis.Cmdable
	interval time.Duration
	rate     int
	service  string
}

func (l *Limiter) Allow() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	allow, err := l.client.Eval(ctx, luaFixWindow, []string{l.service},
		l.interval.Milliseconds(), l.rate).Bool()
	if err != nil {
		return false
	}

	return allow
}

func (l *Limiter) Close() {
	//TODO implement me
}

func NewLimiter(client redis.Cmdable, interval time.Duration, rate int, service string) *Limiter {
	return &Limiter{
		client:   client,
		interval: interval,
		rate:     rate,
		service:  service,
	}
}
