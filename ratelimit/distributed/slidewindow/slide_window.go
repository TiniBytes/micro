package slidewindow

import (
	_ "embed"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"micro/ratelimit/distributed/fixwindow"
	"time"
)

//go:embed lua/slide_window.lua
var luaSlideWindow string

type SlideWindowLimiter struct {
	client   redis.Cmdable
	interval time.Duration
	rate     int
	service  string
}

func NewSlideWindowLimiter(client redis.Cmdable, interval time.Duration, rate int, service string) *fixwindow.Limiter {
	return &fixwindow.Limiter{
		client:   client,
		interval: interval,
		rate:     rate,
		service:  service,
	}
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 不同限流粒度
		allow, err := s.allow(ctx)
		if err != nil {
			return nil, err
		}
		if !allow {
			err = errors.New("rate-limit")
			return
		}

		resp, err = handler(ctx, req)
		return
	}
}

func (s *SlideWindowLimiter) allow(ctx context.Context) (bool, error) {
	return s.client.Eval(ctx, luaSlideWindow, []string{s.service},
		s.interval.Milliseconds(), s.rate, time.Now().UnixMilli()).Bool()
}
