package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"micro/route"
	"sync/atomic"
)

type Balancer struct {
	connections []balancer.SubConn
	index       int32
	len         int32
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	idx := atomic.AddInt32(&b.index, 1)
	conn := b.connections[idx%b.len]

	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	Filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &Balancer{
		connections: connections,
		index:       -1,
		len:         int32(len(connections)),
	}
}
