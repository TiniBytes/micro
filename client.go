package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"micro/registry"
	"time"
)

type Client struct {
	insecure bool
	registry registry.Registry
	timeout  time.Duration
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

func (c *Client) Dial(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.registry != nil {
		builder, err := NewRegistryBuilder(c.registry, c.timeout)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithResolvers(builder))
	}

	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", serviceName), opts...)

	return cc, err
}
