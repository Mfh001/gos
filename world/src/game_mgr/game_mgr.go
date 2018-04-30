package game_mgr

import (
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"google.golang.org/grpc/reflection"
	"goslib/logger"
	"gosconf"
)

type gameAppMgr struct{
}

func Start() {
	startGameDispatcher()
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
	info, err := dispatchGame(in.AccountId, in.ServerId, in.SceneId)
	if err != nil {
		return  nil, err
	}

	return  &pb.DispatchGameReply{
		GameAppId: info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
		SceneId: info.SceneId,
	}, err
}

func (s *gameAppMgr) ReportGameInfo(ctx context.Context, in *pb.ReportGameRequest) (*pb.ReportGameReply, error) {
	reportGameInfo(in.Uuid, in.Host, in.Port, in.Ccu)
	return &pb.ReportGameReply{
		Success: true,
	}, nil
}
