package hash

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"hash/crc32"
)

type Balancer struct {
	connections []balancer.SubConn
	len         int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.len == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	hashCode := info.Ctx.Value("hash_code").([]byte)
	idx := int(crc32.ChecksumIEEE(hashCode))

	conn := b.connections[idx]
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