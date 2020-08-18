package sd

import (
	"io"

	"github.com/go-kit/kit/endpoint"
)

// 根据实例生成Endpoint
type Factory func(instance string) (endpoint.Endpoint, io.Closer, error)
