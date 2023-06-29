package tokenbucket

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"micro/demo/proto"
	"micro/ratelimit"
	"testing"
	"time"
)

func TestLimiter_BuildServerInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		b        func() *Limiter
		ctx      context.Context
		handler  func(ctx context.Context, req interface{}) (interface{}, error)
		wantResp any
		wantErr  error
	}{
		{
			name: "case1_err",
			b: func() *Limiter {
				closeChan := make(chan struct{})
				close(closeChan)
				return &Limiter{
					tokens: make(chan struct{}),
					close:  closeChan,
				}
			},
			ctx:     context.Background(),
			wantErr: errors.New("限流关闭"),
		},
		{
			name: "case2_context_cancel",
			b: func() *Limiter {
				closeChan := make(chan struct{})
				close(closeChan)
				return &Limiter{
					tokens: make(chan struct{}),
					close:  closeChan,
				}
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
		{
			name: "case2_get_token",
			b: func() *Limiter {
				ch := make(chan struct{}, 1)
				ch <- struct{}{}
				return &Limiter{
					tokens: ch,
					close:  make(chan struct{}),
				}
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &proto.GetByIDResp{}, errors.New("mock error")
			},
			ctx:      context.Background(),
			wantResp: &proto.GetByIDResp{},
			wantErr:  errors.New("mock error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := ratelimit.BuildServerInterceptor(tt.b())
			resp, err := interceptor(tt.ctx, proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, tt.handler)
			assert.Equal(t, tt.wantErr.Error(), err.Error())
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantResp, resp)
		})
	}
}

func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(10, 2*time.Second)
	defer func() {
		limiter.Close()
	}()

	interceptor := ratelimit.BuildServerInterceptor(limiter)
	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return &proto.GetByIDResp{}, nil
	}

	resp, err := interceptor(context.Background(), proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Equal(t, &proto.GetByIDResp{}, resp)

	// 触发限流
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	resp, err = interceptor(ctx, proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t, errors.New("rate-limit").Error(), err.Error())
	require.Nil(t, resp)
}
