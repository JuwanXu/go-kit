package service

import (
	"context"
	"errors"

	"github.com/go-kit/kit/example/user/proto"
)

type IUserService interface {
	Login(ctx context.Context, req *proto.LoginReq) (*proto.LoginRes, error)
}

type UserService struct {
}

func NewUserService() IUserService {
	return &UserService{}
}

func (s *UserService) Login(ctx context.Context, req *proto.LoginReq) (*proto.LoginRes, error) {
	if req.Account != "test" {
		return nil, errors.New("account is wrong")
	}
	return &proto.LoginRes{Token: "test"}, nil
}
