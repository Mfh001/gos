package game_mgr

import (
	"gen/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gosconf"
	"goslib/logger"
	"net"
)

var rpcServer *grpc.Server

type gameAppMgr struct {
}

func Start() {
	go startGameAppMgrRPC()
}

func Stop() {
	rpcServer.GracefulStop()
}

func startGameAppMgrRPC() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	lis, err := net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	if err != nil {
		logger.ERR("failed to listen: ", err)
		panic(err)
	}
	logger.INFO("GameAppMgr lis: ", conf.ListenNet, " port: ", conf.ListenPort)

	rpcServer = grpc.NewServer()
	proto.RegisterGameDispatcherServer(rpcServer, &gameAppMgr{})
	reflection.Register(rpcServer)
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *gameAppMgr) DispatchGame(ctx context.Context, in *proto.DispatchGameRequest) (*proto.DispatchGameReply, error) {
	info, err := DispatchGame(in.AccountId, in.SceneId)
	if err != nil {
		return nil, err
	}

	return &proto.DispatchGameReply{
		GameAppId:   info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
	}, err
}
