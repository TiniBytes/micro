package etcd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"micro"
	"micro/demo/grpc/proto"
	"micro/middleware"
	"testing"
)

func TestRegistryEtcd(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	registry, err := NewRegistry(etcdClient)
	require.NoError(t, err)

	us := &UserServer{}

	server, err := micro.NewServer("user-service",
		micro.ServerWithRegister(registry),
		micro.ServerWithGroup("A"),
		micro.ServerWithMiddleware(func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, info interface{}) (interface{}, error) {
				fmt.Println("before")
				reply, err := handler(ctx, info)
				fmt.Println("after")
				return reply, err
			}
		}),
	)
	require.NoError(t, err)
	proto.RegisterUserServiceServer(server, us)

	// 调用start，以为us准备好
	err = server.Start("127.0.0.1:8080")

	t.Log(err)
}

type UserServer struct {
	group string
	proto.UnimplementedUserServiceServer
}

func (s *UserServer) GetByID(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	fmt.Println(request)
	fmt.Println(s.group)
	return &proto.Response{
		User: &proto.User{
			Id:   1,
			Name: "hello,world",
		},
	}, nil
}
