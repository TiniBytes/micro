package etcd

import (
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"micro"
	"micro/demo/grpc/proto"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// 服务注册
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	registry, err := NewRegistry(etcdClient)
	require.NoError(t, err)

	// 服务发现
	client, err := micro.NewClient(micro.ClientInsecure(), micro.ClientWithRegistry(registry, 3*time.Second))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	// 用户服务
	uc := proto.NewUserServiceClient(cc)
	resp, err := uc.GetByID(ctx, &proto.Request{Id: 123456})
	require.NoError(t, err)
	t.Log(resp)
}
