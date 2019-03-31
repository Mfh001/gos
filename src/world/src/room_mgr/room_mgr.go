package room_mgr

import (
	"context"
	"gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gosconf"
	"goslib/logger"
	"net"
)

type roomMgr struct {
}

func Start() {
	startGameAppMgrRPC()
}

func startGameAppMgrRPC() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	lis, err := net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}
	rpcServer := grpc.NewServer()
	proto.RegisterGameDispatcherServer(rpcServer, &roomMgr{})
	reflection.Register(rpcServer)
	logger.INFO("GameAppMgr started!")
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *roomMgr) DispatchGame(ctx context.Context, in *proto.DispatchGameRequest) (*proto.DispatchGameReply, error) {
	info, err := dispatchGame(in.AccountId, in.SceneId)
	if err != nil {
		return nil, err
	}

	return &proto.DispatchGameReply{
		GameAppId:   info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
	}, err
}
