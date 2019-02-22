package agent_mgr

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/logger"
)

type connectAppMgr struct {
}

func Start() {
	startAgentDispatcher()
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
	logger.INFO("ConnectAppMgr started!")
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *connectAppMgr) DispatchPlayer(ctx context.Context, in *pb.DispatchRequest) (*pb.DispatchReply, error) {
	appId, host, port, err := dispatchAgent(in.AccountId, in.GroupId)

	return &pb.DispatchReply{
		ConnectAppId:   appId,
		ConnectAppHost: host,
		ConnectAppPort: port,
	}, err
}

func (s *connectAppMgr) ReportAgentInfo(ctx context.Context, in *pb.AgentInfo) (*pb.OkReply, error) {
	handleReportAgentInfo(in.Uuid, in.Host, in.Port, in.Ccu)
	return &pb.OkReply{Success: true}, nil
}
