package GameAppMgr

import (
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"google.golang.org/grpc/reflection"
	"GameAppMgr/gameApp"
	"goslib/logger"
	"gosconf"
)

// server is used to implement helloworld.GreeterServer.
type gameAppMgr struct{

}

func Start() {
	gameApp.Start()
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
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *gameAppMgr) DispatchGame(ctx context.Context, in *pb.DispatchGameRequest) (*pb.DispatchGameReply, error) {
	info, err := gameApp.Dispatch(in.AccountId, in.ServerId, in.SceneId)

	return  &pb.DispatchGameReply{
		GameAppId: info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
		SceneId: info.SceneId,
	}, err
}
