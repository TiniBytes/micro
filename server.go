package micro

import (
	"context"
	"google.golang.org/grpc"
	"micro/registry"
	"net"
	"time"
)

type Server struct {
	name            string
	registry        registry.Registry
	registryTimeout time.Duration
	listener        net.Listener
	*grpc.Server
}

type ServerOption func(server *Server)

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registryTimeout: time.Second * 10,
	}

	// 函数选项
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// Start 调用start任务服务准备好, 开始注册
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// 有注册中心
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()

		err = s.registry.Register(ctx, &registry.ServiceInstance{
			Name:    s.name,
			Address: listener.Addr().String(),
		})
		if err != nil {
			return err
		}
	}

	// 启动服务
	return s.Serve(listener)
}

func (s *Server) Close() error {
	if s.registry != nil {
		// 服务有注册中心，先从注册中心将服务摘掉
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}

	// 关闭服务
	s.GracefulStop()
	return nil
}

// ServerWithRegister 配置注册中心
func ServerWithRegister(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}
