package agent_mgr

import (
	"net"

	"gen/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	proto.RegisterDispatcherServer(rpcServer, &connectAppMgr{})
	reflection.Register(rpcServer)
	logger.INFO("ConnectAppMgr started!")
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *connectAppMgr) DispatchPlayer(ctx context.Context, in *proto.DispatchRequest) (*proto.DispatchReply, error) {
	appId, host, port, err := dispatchAgent(in.AccountId, in.GroupId)

	return &proto.DispatchReply{
		ConnectAppId:   appId,
		ConnectAppHost: host,
		ConnectAppPort: port,
	}, err
}

func (s *connectAppMgr) ReportAgentInfo(ctx context.Context, in *proto.AgentInfo) (*proto.OkReply, error) {
	handleReportAgentInfo(in.Uuid, in.Host, in.Port, in.Ccu)
	return &proto.OkReply{Success: true}, nil
}
