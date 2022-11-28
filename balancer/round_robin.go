package balancer

import (
	"grpclb/common"
	"math/rand"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// 轮询负载均衡器
const RoundRobin = "round_robin_x"

type roundRobinPickerBuilder struct{}

type roundRobinPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
	next     int
}

func newRoundRobinBuilder() balancer.Builder {
	return base.NewBalancerBuilder(RoundRobin, &roundRobinPickerBuilder{}, base.Config{HealthCheck: true})
}
func init() {
	//注册负载均衡器
	balancer.Register(newRoundRobinBuilder())
}

func (*roundRobinPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	if len(buildInfo.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var scs []balancer.SubConn
	for subConn, subConnInfo := range buildInfo.ReadySCs {
		weight := common.GetWeight(subConnInfo.Address)
		for i := 0; i < weight; i++ {
			scs = append(scs, subConn)
		}
	}

	return &roundRobinPicker{
		subConns: scs,
		next:     rand.Intn(len(scs)),
	}

}

func (p *roundRobinPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	ret := balancer.PickResult{}
	p.mu.Lock()
	ret.SubConn = p.subConns[p.next]
	p.next = (p.next + 1) % len(p.subConns)
	p.mu.Unlock()
	return ret, nil
}
