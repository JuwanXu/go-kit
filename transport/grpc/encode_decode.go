package grpc

import (
	"context"
)

// 解码请求体的函数类型
type DecodeRequestFunc func(context.Context, interface{}) (request interface{}, err error)

// 编码请求体的函数类型
type EncodeRequestFunc func(context.Context, interface{}) (request interface{}, err error)

// 编码响应体的函数类型
type EncodeResponseFunc func(context.Context, interface{}) (response interface{}, err error)

// 解码响应体的函数类型
type DecodeResponseFunc func(context.Context, interface{}) (response interface{}, err error)
