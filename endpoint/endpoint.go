package endpoint

import (
	"context"
)

// 一个Endpoint代表一个Rpc方法
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Nop是一个返回空的Endpoint
func Nop(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }

// Endpoint的中间件
type Middleware func(Endpoint) Endpoint

// 帮助构建中间件
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- {
			next = others[i](next)
		}
		return outer(next)
	}
}

// response可以实现这个接口，go-kit会拦截此异常作为业务异常
type Failer interface {
	Failed() error
}
