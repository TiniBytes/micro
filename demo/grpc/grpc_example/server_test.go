package grpc_example

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"micro/demo/grpc/proto"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	us := &Server{}
	server := grpc.NewServer()
	proto.RegisterUserServiceServer(server, us)

	listen, err := net.Listen("tcp", ":8080")
	require.NoError(t, err)
	err = server.Serve(listen)
	t.Log(err)
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
