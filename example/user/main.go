package main

import (
	"fmt"
	"net"
	"os"

	gt "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/example/user/endpoint/user"
	"github.com/go-kit/kit/example/user/proto"
	"github.com/go-kit/kit/example/user/service"
	tu "github.com/go-kit/kit/example/user/transport/user"
)

func main() {
	u := &service.UserService{}
	e := user.MakeLoginEndPoint(u)
	t := tu.New(e)

	l, err := net.Listen("tcp", ":8881")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(gt.Interceptor))
	proto.RegisterUserServer(s, t)
	if err = s.Serve(l); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
