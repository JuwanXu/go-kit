package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
)

/*
	type UserTransport struct {
		grpc.Handler
	}

	func New(endpoint endpoint.Endpoint) proto.UserServer {
		return &UserTransport{grpc.NewServer(endpoint, DecodeLoginReq, EncodeLoginRes)}
	}

	func (t *UserTransport) Login(ctx context.Context, req *proto.LoginReq) (*proto.LoginRes, error) {
		_, res, err := t.ServeGRPC(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.(*proto.LoginRes), err
	}

	此处handler应该被Transport所用，定义方法之后，调用ServeGRPC方法
*/
type Handler interface {
	ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error)
}

type ServerFinalizerFunc func(ctx context.Context, err error)

// Server封装了一个Endpoint
// 即一个Server对应一个对外暴露的rpc方法
type Server struct {
	e            endpoint.Endpoint      // 方法的真正实现
	dec          DecodeRequestFunc      // 请求解码
	enc          EncodeResponseFunc     // 响应编码
	before       []ServerRequestFunc    // 前置处理器
	after        []ServerResponseFunc   // 后置处理器
	finalizer    []ServerFinalizerFunc  // 最终置处理器
	errorHandler transport.ErrorHandler // 异常处理器
}

// 初始化一个Server
func NewServer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ServerOption,
) *Server {
	// 默认的异常处理器为日志打印处理器
	s := &Server{
		e:            e,
		dec:          dec,
		enc:          enc,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

type ServerOption func(*Server)

func ServerBefore(before ...ServerRequestFunc) ServerOption {
	return func(s *Server) { s.before = append(s.before, before...) }
}

func ServerAfter(after ...ServerResponseFunc) ServerOption {
	return func(s *Server) { s.after = append(s.after, after...) }
}

func ServerErrorHandler(errorHandler transport.ErrorHandler) ServerOption {
	return func(s *Server) { s.errorHandler = errorHandler }
}

func ServerFinalizer(f ...ServerFinalizerFunc) ServerOption {
	return func(s *Server) { s.finalizer = append(s.finalizer, f...) }
}

// 核心方法
func (s Server) ServeGRPC(ctx context.Context, req interface{}) (retctx context.Context, resp interface{}, err error) {
	// 取出元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	// 调用最终执行器
	if len(s.finalizer) > 0 {
		defer func() {
			for _, f := range s.finalizer {
				f(ctx, err)
			}
		}()
	}
	// 调用前置处理器
	for _, f := range s.before {
		ctx = f(ctx, md)
	}
	var (
		request  interface{}
		response interface{}
		grpcResp interface{}
	)
	// 解码请求
	request, err = s.dec(ctx, req)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}
	// 调用方法处理请求
	response, err = s.e(ctx, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}
	// 调用最终处理器
	var mdHeader, mdTrailer metadata.MD
	for _, f := range s.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}
	// 编码响应
	grpcResp, err = s.enc(ctx, response)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return ctx, nil, err
	}
	if len(mdHeader) > 0 {
		if err = grpc.SendHeader(ctx, mdHeader); err != nil {
			s.errorHandler.Handle(ctx, err)
			return ctx, nil, err
		}
	}
	if len(mdTrailer) > 0 {
		if err = grpc.SetTrailer(ctx, mdTrailer); err != nil {
			s.errorHandler.Handle(ctx, err)
			return ctx, nil, err
		}
	}
	return ctx, grpcResp, nil
}

// 拦截器，将方法名注入ctx，以供go-kit使用
func Interceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	ctx = context.WithValue(ctx, ContextKeyRequestMethod, info.FullMethod)
	return handler(ctx, req)
}
