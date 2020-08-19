package lb

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// 定义一个负载均衡接口
type Balancer interface {
	Endpoint() (endpoint.Endpoint, error)
}

// ErrNoEndpoints is returned when no qualifying endpoints are available.
var ErrNoEndpoints = errors.New("no endpoints available")
