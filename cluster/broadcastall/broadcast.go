package broadcast

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"micro/registry"
	"reflect"
	"sync"
)

// ClusterBuilder 广播（全部响应）
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
		ok, ch := isBroadCast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		defer func() {
			close(ch)
		}()

		instances, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		typ := reflect.TypeOf(reply).Elem()
		wg.Add(len(instances))
		for _, ins := range instances {
			addr := ins.Address

			// 并发调用每一个节点
			go func() {
				var clientConn *grpc.ClientConn
				clientConn, err = grpc.Dial(addr, b.options...)
				if err != nil {
					ch <- Response{Err: err}
					wg.Done()
					return
				}

				newReply := reflect.New(typ).Interface()
				err = invoker(ctx, method, req, newReply, clientConn, opts...)
				select {
				case <-ctx.Done():
					err = fmt.Errorf("response not received, %w", ctx.Err())
				case ch <- Response{Reply: newReply, Err: err}:
				}

				wg.Done()
			}()
		}
		wg.Wait()
		return err
	}
}

func UseBroadCast(ctx context.Context) (context.Context, <-chan Response) {
	ch := make(chan Response)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct{}

func isBroadCast(ctx context.Context) (bool, chan Response) {
	val, ok := ctx.Value(broadcastKey{}).(chan Response)
	return ok, val
}

// Response 广播响应
type Response struct {
	Reply any
	Err   error
}
