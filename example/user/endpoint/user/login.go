package user

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/go-kit/kit/example/user/proto"
	"github.com/go-kit/kit/example/user/service"
)

func MakeLoginEndPoint(s service.IUserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*proto.LoginReq)
		return s.Login(ctx, req)
	}
}
