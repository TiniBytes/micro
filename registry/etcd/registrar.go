package etcd

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"golang.org/x/net/context"
	"micro/registry"
	"sync"
)

type Registry struct {
	client  *clientv3.Client
	session *concurrency.Session
	cancels []func()
	mutex   sync.Mutex
}

func NewRegistry(client *clientv3.Client) (*Registry, error) {
	// session内部已经实现心跳
	session, err := concurrency.NewSession(client, concurrency.WithTTL(60))
	if err != nil {
		return nil, err
	}
	return &Registry{
		client:  client,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	val, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 将服务实例和租约信息写入etcd
	_, err = r.client.Put(ctx, r.instanceKey(service), string(val), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, service *registry.ServiceInstance) error {
	_, err := r.client.Delete(ctx, r.instanceKey(service))
	return err
}

func (r *Registry) ListServices(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	response, err := r.client.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	res := make([]*registry.ServiceInstance, 0, len(response.Kvs))
	for _, kv := range response.Kvs {
		si := &registry.ServiceInstance{}

		err = json.Unmarshal(kv.Value, si)
		if err != nil {
			return nil, err
		}

		res = append(res, si)
	}

	return res, nil
}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()

	ctx = clientv3.WithRequireLeader(ctx)
	watchResp := r.client.Watch(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())

	res := make(chan registry.Event)
	go func() {
		for {
			select {
			case resp := <-watchResp:
				// 监听到事件变更
				if resp.Err() != nil || resp.Canceled {
					return
				}

				for range resp.Events {
					res <- registry.Event{}
				}
			case <-ctx.Done():
				// 退出信号
				return
			}
		}
	}()

	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()

	// 逐个关闭监听事件
	for _, cancel := range cancels {
		cancel()
	}

	return r.session.Close()
}

func (r *Registry) instanceKey(service *registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", service.Name, service.Address)
}

func (r *Registry) serviceKey(service string) string {
	return fmt.Sprintf("/micro/%s", service)
}
