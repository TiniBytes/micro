package broadcast

import (
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"micro"
	"micro/demo/grpc/proto"
	"time"

	"micro/registry/etcd"
	"testing"
)

func TestUseBroadCast(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	registry, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	var servers []*UserServer
	for i := 0; i < 3; i++ {
		server, err := micro.NewServer("user-service", micro.ServerWithRegister(registry))
		require.NoError(t, err)
		us := &UserServer{
			index: i,
			group: "",
		}
		servers = append(servers, us)
		proto.RegisterUserServiceServer(server, us)
		port := fmt.Sprintf("127.0.0.1:808%d", i)
		eg.Go(func() error {
			return server.Start(port)
		})
		//defer func() {
		//	_ = server.Close()
		//}()
	}
	time.Sleep(3 * time.Second)

	// client
	client, err := micro.NewClient(
		micro.ClientInsecure(),
		micro.ClientWithRegistry(registry, 10*time.Second),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//ctx = context.WithValue(ctx, "group", "A")
	ctx, respCh := UseBroadCast(ctx)
	go func() {
		res := <-respCh
		fmt.Println(res)
	}()

	// 集群广播
	bd := NewClusterBuilder(registry, "user-service", grpc.WithInsecure())
	cc, err := client.Dial(ctx, "user-service", grpc.WithUnaryInterceptor(bd.BuildUnaryInterceptor()))
	require.NoError(t, err)

	// 用户服务
	uc := proto.NewUserServiceClient(cc)
	resp, err := uc.GetByID(ctx, &proto.Request{Id: 123456})
	require.NoError(t, err)
	t.Log(resp)

}

type UserServer struct {
	index int
	group string
	proto.UnimplementedUserServiceServer
}

func (s *UserServer) GetByID(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	s.index++
	fmt.Println(request)
	fmt.Println(s.group)
	return &proto.Response{
		User: &proto.User{
			Id:   1,
			Name: fmt.Sprintf("hello,world %d", s.index),
		},
	}, nil
}
