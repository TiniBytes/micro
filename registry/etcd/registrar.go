package etcd

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"golang.org/x/net/context"
	"micro/registry"
)

type Registry struct {
	Client  *clientv3.Client
	session *concurrency.Session
}

func NewRegistry(client *clientv3.Client) (*Registry, error) {
	// session内部已经实现心跳
	session, err := concurrency.NewSession(client, concurrency.WithTTL(60))
	if err != nil {
		return nil, err
	}
	return &Registry{
		Client:  client,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	val, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 将服务实例和租约信息写入etcd
	_, err = r.Client.Put(ctx, r.instanceKey(service), string(val), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, service *registry.ServiceInstance) error {
	_, err := r.Client.Delete(ctx, r.instanceKey(service))
	return err
}

func (r *Registry) ListServices(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Registry) Close() error {
	return r.session.Close()
}

func (r *Registry) instanceKey(service *registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", service.Name, service.Address)
}

func (r *Registry) serviceKey(service *registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s", service.Name)
}
