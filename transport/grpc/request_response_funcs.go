package grpc

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

const (
	ContextKeyRequestMethod contextKey = iota
)

type contextKey int

// 客户端前置执行器函数类型
type ClientRequestFunc func(context.Context, *metadata.MD) context.Context

// 客户端后置执行器函数类型
type ClientResponseFunc func(ctx context.Context, header metadata.MD, trailer metadata.MD) context.Context

// 服务端前置执行器函数类型
type ServerRequestFunc func(context.Context, metadata.MD) context.Context

// 服务端后置执行器函数类型
type ServerResponseFunc func(ctx context.Context, header *metadata.MD, trailer *metadata.MD) context.Context

func SetRequestHeader(key, val string) ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func SetResponseHeader(key, val string) ServerResponseFunc {
	return func(ctx context.Context, md *metadata.MD, _ *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func SetResponseTrailer(key, val string) ServerResponseFunc {
	return func(ctx context.Context, _ *metadata.MD, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

func EncodeKeyValue(key, val string) (string, string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, binHdrSuffix) {
		val = base64.StdEncoding.EncodeToString([]byte(val))
	}
	return key, val
}
