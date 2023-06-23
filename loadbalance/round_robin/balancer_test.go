package round_robin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"micro/demo/grpc/proto"
	"net"
	"testing"
	"time"
)

func TestBuilder_Pick(t *testing.T) {
	tests := []struct {
		name              string
		b                 *Balancer
		wantErr           error
		want              balancer.PickResult
		wantSubConn       SubConn
		wantBalancerIndex int32
	}{
		{
			name: "case1",
			b: &Balancer{
				connections: []balancer.SubConn{
					SubConn{name: "127.0.0.1:8080"},
					SubConn{name: "127.0.0.1:8081"},
				},
				index: -1,
				len:   2,
			},
			wantSubConn:       SubConn{name: "127.0.0.1:8080"},
			wantBalancerIndex: 0,
		},
		{
			name: "case2",
			b: &Balancer{
				connections: []balancer.SubConn{
					SubConn{name: "127.0.0.1:8080"},
					SubConn{name: "127.0.0.1:8081"},
				},
				index: 1,
				len:   2,
			},
			wantSubConn:       SubConn{name: "127.0.0.1:8080"},
			wantBalancerIndex: 2,
		},
		{
			name: "case3",
			b: &Balancer{
				connections: []balancer.SubConn{},
				index:       -1,
			},
			wantErr: balancer.ErrNoSubConnAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.b.Pick(balancer.PickInfo{})
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantSubConn.name, result.SubConn.(SubConn).name)
			assert.NotNil(t, result.Done)
			assert.Equal(t, tt.wantBalancerIndex, tt.b.index)
		})
	}
}

type SubConn struct {
	balancer.SubConn
	name string
}

func TestBalancer_Pick(t *testing.T) {
	// 服务端
	go func() {
		us := &Server{}
		server := grpc.NewServer()
		proto.RegisterUserServiceServer(server, us)

		listen, err := net.Listen("tcp", ":8080")
		require.NoError(t, err)
		err = server.Serve(listen)
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	// 客户端

	balancer.Register(base.NewBalancerBuilder("DOME_ROUND_ROBIN", &Builder{}, base.Config{
		HealthCheck: true,
	}))

	cc, err := grpc.Dial("localhost:8080", grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy": "DOME_ROUND_ROBIN"}`))
	require.NoError(t, err)

	client := proto.NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.GetByID(ctx, &proto.Request{Id: 123456})
	require.NoError(t, err)
	t.Log(resp)
}

type Server struct {
	proto.UnimplementedUserServiceServer
}

func (s *Server) GetByID(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	fmt.Println(request)
	return &proto.Response{
		User: &proto.User{
			Id:   1,
			Name: "hello,world",
		},
	}, nil
}
