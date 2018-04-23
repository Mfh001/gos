package GameAppMgr

import (
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "connectAppProto"
	"google.golang.org/grpc/reflection"
	"GameAppMgr/gameApp"
	"goslib/logger"
)

const (
	port = ":50052"
)

// server is used to implement helloworld.GreeterServer.
type gameAppMgr struct{

}

func Start() {
	gameApp.Start()
	startGameAppMgrRPC()
}

func startGameAppMgrRPC() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}
	rpcServer := grpc.NewServer()
	pb.RegisterGameDispatcherServer(rpcServer, &gameAppMgr{})
	// Register reflection service on gRPC server.
	reflection.Register(rpcServer)
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *gameAppMgr) DispatchGame(ctx context.Context, in *pb.DispatchGameRequest) (*pb.DispatchGameReply, error) {
	host, port, err := gameApp.Dispatch(in.AccountId, in.ServerId, in.SceneId)

	return  &pb.DispatchGameReply{
		GameAppHost: host,
		GameAppPort: port,
	}, err
}
