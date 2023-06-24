package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"micro/route"
	"sync/atomic"
)

type Balancer struct {
	connections []*subConn
	index       int32
	len         int32
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// filter
	candidates := make([]*subConn, 0, b.len)
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}

	// load balancer
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	idx := atomic.AddInt32(&b.index, 1)
	conn := candidates[int(idx)%len(candidates)]

	return balancer.PickResult{
		SubConn: conn.conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	Filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*subConn, 0, len(info.ReadySCs))

	for c, ci := range info.ReadySCs {
		connections = append(connections, &subConn{
			conn: c,
			addr: ci.Address,
		})
	}

	return &Balancer{
		connections: connections,
		index:       -1,
		len:         int32(len(connections)),
		filter:      b.Filter,
	}
}

type subConn struct {
	conn balancer.SubConn
	addr resolver.Address
}
