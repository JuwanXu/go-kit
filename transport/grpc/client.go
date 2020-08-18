package grpc

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
)

// 最终执行器，可以用来记录、打印异常等
type ClientFinalizerFunc func(ctx context.Context, err error)

// 封装gRpc原装的client
type Client struct {
	client      *grpc.ClientConn      // gRpc连接
	serviceName string                // 服务名称
	method      string                // 方法
	enc         EncodeRequestFunc     // 编码请求
	dec         DecodeResponseFunc    // 解码响应
	grpcReply   reflect.Type          // 响应结构体
	before      []ClientRequestFunc   // 前置执行器
	after       []ClientResponseFunc  // 后置执行器
	finalizer   []ClientFinalizerFunc //
}

// 创建一个go-kit gRpc客户端
func NewClient(
	cc *grpc.ClientConn,
	serviceName string,
	method string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	grpcReply interface{},
	options ...ClientOption,
) *Client {
	c := &Client{
		client: cc,
		method: fmt.Sprintf("/%s/%s", serviceName, method),
		enc:    enc,
		dec:    dec,
		// 保证用户传入响应结构体或者其指针均可
		grpcReply: reflect.TypeOf(
			reflect.Indirect(
				reflect.ValueOf(grpcReply),
			).Interface(),
		),
		before: []ClientRequestFunc{},
		after:  []ClientResponseFunc{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// 配置函数
type ClientOption func(*Client)

// 配置前置执行器
func ClientBefore(before ...ClientRequestFunc) ClientOption {
	return func(c *Client) { c.before = append(c.before, before...) }
}

// 配置后置执行器
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(c *Client) { c.after = append(c.after, after...) }
}

// 配置最终执行器
func ClientFinalizer(f ...ClientFinalizerFunc) ClientOption {
	return func(s *Client) { s.finalizer = append(s.finalizer, f...) }
}

// 返回客户端对应的Endpoint
// 用户侧需要再封装Endpoint，继承service接口，然后调用此Endpoint即可
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// 根context是在用户调用方法时传入(也可能非根节点)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		// 若最终执行器不为空，则执行
		if c.finalizer != nil {
			defer func() {
				for _, f := range c.finalizer {
					f(ctx, err)
				}
			}()
		}
		// 向context写入方法名
		ctx = context.WithValue(ctx, ContextKeyRequestMethod, c.method)
		// 编码请求
		req, err := c.enc(ctx, request)
		if err != nil {
			return nil, err
		}
		// 前置执行器生成MD并填充进入context
		md := &metadata.MD{}
		for _, f := range c.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewOutgoingContext(ctx, *md)
		// 调用方法，处理结果
		var header, trailer metadata.MD
		grpcReply := reflect.New(c.grpcReply).Interface()
		if err = c.client.Invoke(
			ctx, c.method, req, grpcReply, grpc.Header(&header),
			grpc.Trailer(&trailer),
		); err != nil {
			return nil, err
		}
		// 调用后置处理器
		for _, f := range c.after {
			ctx = f(ctx, header, trailer)
		}
		// 处理响应
		response, err = c.dec(ctx, grpcReply)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}
