package registry

import (
	"golang.org/x/net/context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, service *ServiceInstance) error
	UnRegister(ctx context.Context, service *ServiceInstance) error
	ListServices(ctx context.Context, name string) ([]*ServiceInstance, error)
	Subscribe(serviceName string) (<-chan Event, error)
	io.Closer
}

type ServiceInstance struct {
	Name    string
	Address string
}

type Event struct{}
