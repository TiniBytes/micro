package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type Balancer struct {
	connections []balancer.SubConn
	len         int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	r := rand.Intn(int(b.len))
	conn := b.connections[r]

	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct{}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}

	return &Balancer{
		connections: connections,
		len:         int32(len(connections)),
	}
}
