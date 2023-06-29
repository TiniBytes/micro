package leakylimiter

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"micro/demo/proto"
	"micro/ratelimit"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(3 * time.Second)
	interceptor := ratelimit.BuildServerInterceptor(limiter)
	//interceptor := NewLimiter(3*time.Second, 1).BuildServerInterceptor()
	cnt := 0
	time.Sleep(3 * time.Second)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return &proto.GetByIDResp{}, nil
	}

	resp, err := interceptor(context.Background(), proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	require.Equal(t, &proto.GetByIDResp{}, resp)

	// 触发限流
	resp, err = interceptor(context.Background(), proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t, errors.New("rate-limit").Error(), err.Error())
	require.Nil(t, resp)

	// 新窗口出现
	time.Sleep(3 * time.Second)
	resp, err = interceptor(context.Background(), proto.GetByIDReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	require.Equal(t, &proto.GetByIDResp{}, resp)
}
