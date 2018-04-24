package ConnectAppMgr

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"google.golang.org/grpc/reflection"
	"ConnectAppMgr/connectApp"
	"gosconf"
	"goslib/logger"
)

// server is used to implement helloworld.GreeterServer.
type connectAppMgr struct{
}

func Start() {
	connectApp.Start()
	startConnectAppMgrRpc()
}

func startConnectAppMgrRpc() {
	conf := gosconf.RPC_FOR_CONNECT_APP_MGR
	lis, err := net.Listen(conf.ListenNet, conf.ListenAddr)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}
	rpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(rpcServer, &connectAppMgr{})
	reflection.Register(rpcServer)
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *connectAppMgr) DispatchPlayer(ctx context.Context, in *pb.DispatchRequest) (*pb.DispatchReply, error) {
	appId, host, port, err := connectApp.Dispatch(in.AccountId, in.GroupId)

	return  &pb.DispatchReply{
		ConnectAppId: appId,
		ConnectAppHost: host,
		ConnectAppPort: port,
	}, err
}
