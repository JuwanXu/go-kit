package lb

import (
	"math/rand"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
)

// 随机返回一个可用的Endpoint
func NewRandom(s sd.Endpointer, seed int64) Balancer {
	return &random{
		s: s,
		r: rand.New(rand.NewSource(seed)),
	}
}

type random struct {
	s sd.Endpointer
	r *rand.Rand
}

func (r *random) Endpoint() (endpoint.Endpoint, error) {
	endpoints, err := r.s.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}
	return endpoints[r.r.Intn(len(endpoints))], nil
}
