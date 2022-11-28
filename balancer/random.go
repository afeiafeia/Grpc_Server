package balancer

import (
	"grpclb/common"
	"math/rand"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// 随机负载均衡器
const Random = "random_x"

type randomPickerBuilder struct{}

type randomPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
	next     int
}

func newRandommBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Random, &roundRobinPickerBuilder{}, base.Config{HealthCheck: true})
}
func init() {
	//注册负载均衡器
	balancer.Register(newRandommBuilder())
}

func (*randomPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
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

func (p *randomPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	ret := balancer.PickResult{}
	p.mu.Lock()
	idx := rand.Int() % len(p.subConns)
	ret.SubConn = p.subConns[idx]
	p.mu.Unlock()
	return ret, nil
}
