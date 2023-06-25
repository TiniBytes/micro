package broadcast

import (
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"micro/registry"
)

type ClusterBuilder struct {
	registry registry.Registry
	service  string
	options  []grpc.DialOption
}

func NewClusterBuilder(r registry.Registry, service string, options ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry: r,
		service:  service,
		options:  options,
	}
}

func (b ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBroadCast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		instances, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}

		var eg errgroup.Group
		for _, ins := range instances {
			addr := ins.Address

			// 并发调用每一个节点
			eg.Go(func() error {
				var clientConn *grpc.ClientConn
				clientConn, err = grpc.Dial(addr, b.options...)
				if err != nil {
					return err
				}

				err = invoker(ctx, method, req, reply, clientConn, opts...)
				return err
			})
		}
		return eg.Wait()
	}
}

func UseBroadCast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

type broadcastKey struct{}

func isBroadCast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}
