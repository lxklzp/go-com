package grpcc

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/internal/grpcs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var Client client

type client struct {
	conn *grpc.ClientConn
	Cli  proto.AppClient
}

func (c *client) Connect() {
	c.conn, _ = grpc.Dial(config.C.App.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(c.unaryInterceptor))
	c.Cli = proto.NewAppClient(c.conn)

	// 接口调用示例
	fmt.Println(c.Cli.HeheAdd(context.TODO(), &proto.HeheReqData{
		Id:     0,
		Name:   "ccc",
		Age:    10,
		UserId: 0,
	}))
}

// 注入认证信息token
func (c *client) unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.Pairs("token", config.C.App.GrpcToken)
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

func (c *client) Close() {
	c.conn.Close()
}
