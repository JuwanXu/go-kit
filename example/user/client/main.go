package main

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	gt "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/example/user/proto"
	"github.com/go-kit/kit/example/user/service"
)

type EndPointServer struct {
	LoginEndPoint endpoint.Endpoint
}

func (s EndPointServer) Login(ctx context.Context, in *proto.LoginReq) (*proto.LoginRes, error) {
	res, err := s.LoginEndPoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return res.(*proto.LoginRes), nil
}

func main() {
	conn, err := grpc.Dial("127.0.0.1:8881", grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	svr := NewClient(conn)
	ack, err := svr.Login(context.Background(), &proto.LoginReq{
		Account:  "test",
		Password: "123456",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ack)
}

func NewClient(conn *grpc.ClientConn) service.IUserService {
	e := gt.NewClient(conn,
		"proto.User",
		"Login",
		RequestLogin,
		ResponseLogin,
		proto.LoginRes{}).Endpoint()
	return EndPointServer{LoginEndPoint: e}
}

func RequestLogin(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*proto.LoginReq)
	return &proto.LoginReq{Account: req.Account, Password: req.Password}, nil
}

func ResponseLogin(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*proto.LoginRes)
	return &proto.LoginRes{Token: resp.Token}, nil
}
