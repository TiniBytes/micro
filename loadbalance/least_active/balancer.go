package least_active

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
	"sync/atomic"
)

type Balancer struct {
	connections []*activeConn
	len         int32
	mutex       sync.Mutex
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	res := &activeConn{
		count: math.MaxUint32,
	}
	for _, c := range b.connections {
		if atomic.LoadUint32(&c.count) < res.count {
			res = c
			break
		}
	}

	atomic.AddUint32(&res.count, 1)
	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&res.count, -1)
		},
	}, nil
}

type Builder struct{}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*activeConn, 0, len(info.ReadySCs))

	for c := range info.ReadySCs {
		connections = append(connections, &activeConn{
			conn: c,
		})
	}

	return &Balancer{
		connections: connections,
	}
}

type activeConn struct {
	conn  balancer.SubConn
	count uint32
}
