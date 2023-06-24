package weight_round_robin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestWeightBalancer_Pick(t *testing.T) {
	b := Balancer{
		connections: []*weightConn{
			{
				conn: SubConn{
					name: "weight-5",
				},
				weight:          5,
				efficientWeight: 5,
				currentWeight:   5,
			},
			{
				conn: SubConn{
					name: "weight-4",
				},
				weight:          4,
				efficientWeight: 4,
				currentWeight:   4,
			},
			{
				conn: SubConn{
					name: "weight-3",
				},
				weight:          3,
				efficientWeight: 3,
				currentWeight:   3,
			},
		},
	}

	pick, err := b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-5", pick.SubConn.(SubConn).name)

	pick, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-4", pick.SubConn.(SubConn).name)

	pick, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-3", pick.SubConn.(SubConn).name)

	pick, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-5", pick.SubConn.(SubConn).name)
}

type SubConn struct {
	balancer.SubConn
	name string
}
