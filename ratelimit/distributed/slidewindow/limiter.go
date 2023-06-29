package slidewindow

import (
	_ "embed"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"time"
)

//go:embed lua/slide_window.lua
var luaSlideWindow string

type Limiter struct {
	client   redis.Cmdable
	interval time.Duration
	rate     int
	service  string
}

func (l *Limiter) Allow() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	allow, err := l.client.Eval(ctx, luaSlideWindow, []string{l.service},
		l.interval.Milliseconds(), l.rate, time.Now().UnixMilli()).Bool()
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
