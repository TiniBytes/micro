package micro

import (
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"micro/registry"
	"sync"
	"time"
)

type RegistryBuilder struct {
	registry registry.Registry
	timeout  time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*RegistryBuilder, error) {
	return &RegistryBuilder{
		registry: r,
		timeout:  timeout,
	}, nil
}

func (r *RegistryBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := &RegistryResolver{
		cc:       cc,
		registry: r.registry,
		target:   target,
		timeout:  r.timeout,
	}
	res.ResolveNow(resolver.ResolveNowOptions{})

	go func() {
		res.watch()
	}()

	return res, nil
}

func (r *RegistryBuilder) Scheme() string {
	return "registry"
}

type RegistryResolver struct {
	cc       resolver.ClientConn
	registry registry.Registry
	target   resolver.Target
	timeout  time.Duration
	close    chan struct{}
}

func (r *RegistryResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.resolve()
}

func (r *RegistryResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	instances, err := r.registry.ListServices(ctx, r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	address := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{
			Addr:       si.Address,
			Attributes: attributes.New("group", si.Group),
		})
	}

	err = r.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
}

func (r *RegistryResolver) watch() {
	events, err := r.registry.Subscribe(r.target.Endpoint())
	if err != nil {
		r.cc.ReportError(err)
		return
	}

	// 监听events
	for {
		select {
		case <-events:
			// 服务变更事件
			r.resolve()
		case <-r.close:
			// 退出
			return
		}
	}
}

func (r *RegistryResolver) Close() {
	fn := func() {
		close(r.close)
	}

	once := sync.Once{}
	once.Do(fn)
}
