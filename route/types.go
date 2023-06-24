package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilter struct {
	Croup string
}

func (g GroupFilter) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		target := addr.Attributes.Value("group").(string)
		input := info.Ctx.Value("group").(string)
		return target == input
	}
}
