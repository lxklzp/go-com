package grpcs

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/mod"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/grpcs/proto"
	"go-com/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"net"
	"time"
)

var Server server

type server struct {
	serv *grpc.Server
	proto.UnimplementedAppServer
}

func (s *server) Run() {
	address := config.C.App.GrpcAddr
	listen, _ := net.Listen("tcp", address)

	s.serv = grpc.NewServer(grpc.UnaryInterceptor(s.unaryInterceptorServ))
	proto.RegisterAppServer(s.serv, s)
	if err := s.serv.Serve(listen); err != nil {
		logr.L.Fatal(err)
	}
}

// 验证token
func (s *server) valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	return authorization[0] == config.C.App.GrpcToken
}

// 请求统一处理
func (s *server) unaryInterceptorServ(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (data interface{}, err error) {
	// 接口异常处理
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			msg := tool.ErrorStack(recoverErr)
			err = errors.New(msg)
		}
	}()

	if config.C.App.DebugMode {
		p, _ := peer.FromContext(ctx)
		logr.L.Debugf("[grpc req] %s %s\n%s", info.FullMethod, p.Addr, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "请在metadata中携带token")
	}
	if !s.valid(md["token"]) {
		return nil, status.Errorf(codes.Unauthenticated, "token有误")
	}
	data, err = handler(ctx, req)
	return data, err
}

func (s *server) Stop() {
	s.serv.Stop()
}

/*---------- 接口方法，代码自动补全方式：ctrl+shift+P，输入ServerApp ----------*/

func (s *server) HeheAdd(ctx context.Context, param *proto.HeheReqData) (*proto.HeheRespData, error) {
	var err error
	var m model.Hehe
	paramJson, _ := json.Marshal(param)
	json.Unmarshal(paramJson, &m)

	var mName mod.Name
	app.Pg.Model(m).Where("name=?", m.Name).Take(&mName)
	if mName.Name != "" {
		return nil, errors.New("名称已存在")
	}

	m.CreateTime = tool.Timestamp(time.Now())
	if err = app.Pg.Create(&m).Error; err != nil {
		return nil, err
	}
	return &proto.HeheRespData{Id: m.ID}, nil
}
