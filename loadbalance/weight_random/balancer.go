package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"micro/route"
)

type Balancer struct {
	connections []*weightConn
	totalWeight uint32
	len         int32
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var idx int
	target := rand.Intn(int(b.totalWeight) + 1)
	for i, c := range b.connections {
		target -= int(c.weight)
		if target < 0 {
			idx = i
			break
		}
	}

	conn := b.connections[idx]
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
	connections := make([]*weightConn, 0, len(info.ReadySCs))
	var totalWeight uint32

	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		totalWeight += weight

		connections = append(connections, &weightConn{
			conn:   sub,
			weight: weight,
		})
	}

	return &Balancer{
		connections: connections,
		len:         int32(len(connections)),
		totalWeight: totalWeight,
	}
}

type weightConn struct {
	conn   balancer.SubConn
	weight uint32
}
