package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"micro/registry"
	"time"
)

type Client struct {
	insecure bool
	registry registry.Registry
	timeout  time.Duration
	balancer string
}

type ClientOption func(client *Client)

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}

	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ClientInsecure() ClientOption {
	return func(client *Client) {
		client.insecure = true
	}
}

func ClientWithRegistry(r registry.Registry, timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.registry = r
		client.timeout = timeout
	}
}

func ClientWithPickBuilder(name string, b base.PickerBuilder) ClientOption {
	return func(client *Client) {
		balancer.Register(base.NewBalancerBuilder(name, b, base.Config{
			HealthCheck: true,
		}))
		client.balancer = name
	}
}

func (c *Client) Dial(ctx context.Context, serviceName string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.registry != nil {
		builder, err := NewRegistryBuilder(c.registry, c.timeout)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithResolvers(builder))
	}

	if c.balancer != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, c.balancer)))
	}

	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	if len(options) != 0 {
		opts = append(opts, options...)
	}

	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", serviceName), opts...)

	return cc, err
}
