package grpc_example

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"micro/demo/grpc/proto"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("registry:///localhost:8080", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)

	client := proto.NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.GetByID(ctx, &proto.Request{Id: 123456})
	require.NoError(t, err)
	t.Log(resp)
}
