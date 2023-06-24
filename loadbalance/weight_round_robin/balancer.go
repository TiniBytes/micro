package weight_round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
)

type Balancer struct {
	connections []*weightConn
	mutex       sync.Mutex
}

func (w *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var totalWeight uint32
	var res *weightConn

	for _, c := range w.connections {
		c.mutex.Lock()
		totalWeight += c.efficientWeight
		c.currentWeight += c.efficientWeight

		if res == nil || res.currentWeight < c.currentWeight {
			res = c
		}
		c.mutex.Unlock()
	}
	res.mutex.Lock()
	res.currentWeight -= totalWeight
	res.mutex.Unlock()

	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			w.mutex.Lock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil && res.efficientWeight == math.MaxUint32 {
				return
			}

			if info.Err != nil {
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
			w.mutex.Unlock()
		},
	}, nil
}

type BalancerBuilder struct{}

func (w *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*weightConn, 0, len(info.ReadySCs))

	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(int32)

		// 全部初始化为weight
		connections = append(connections, &weightConn{
			conn:            sub,
			weight:          uint32(weight),
			currentWeight:   uint32(weight),
			efficientWeight: uint32(weight),
		})
	}

	return &Balancer{
		connections: connections,
	}
}

type weightConn struct {
	conn            balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
	mutex           sync.Mutex
}
