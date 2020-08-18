package user

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/grpc"

	"github.com/go-kit/kit/example/user/proto"
)

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

func DecodeLoginReq(_ context.Context, req interface{}) (interface{}, error) {
	innerReq := req.(*proto.LoginReq)
	return &proto.LoginReq{Account: innerReq.GetAccount(), Password: innerReq.GetPassword()}, nil
}

func EncodeLoginRes(_ context.Context, res interface{}) (interface{}, error) {
	innerRes := res.(*proto.LoginRes)
	return &proto.LoginRes{Token: innerRes.Token}, nil
}
