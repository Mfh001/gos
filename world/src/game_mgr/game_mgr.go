package game_mgr

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/logger"
	"net"
)

type gameAppMgr struct {
}

func Start() {
	startGameAppMgrRPC()
}

func startGameAppMgrRPC() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	lis, err := net.Listen(conf.ListenNet, conf.ListenAddr)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}
	rpcServer := grpc.NewServer()
	pb.RegisterGameDispatcherServer(rpcServer, &gameAppMgr{})
	reflection.Register(rpcServer)
	logger.INFO("GameAppMgr started!")
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *gameAppMgr) DispatchGame(ctx context.Context, in *pb.DispatchGameRequest) (*pb.DispatchGameReply, error) {
	info, err := dispatchGame(in.AccountId, in.SceneId)
	if err != nil {
		return nil, err
	}

	return &pb.DispatchGameReply{
		GameAppId:   info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
	}, err
}
