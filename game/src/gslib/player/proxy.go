package player

import (
	"gosconf"
	"net"
	"io"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"goslib/logger"
	"context"
	"sync"
	"gslib/scene_mgr"
	"google.golang.org/grpc/metadata"
	"sync/atomic"
	"strconv"
)

type StreamServer struct {
	ConnectAppId string
}

var accountConnectMap = sync.Map{}
var onlinePlayers int32

func OnlinePlayers() int32 {
	return onlinePlayers
}

var StreamRpcListenPort string
func StartRpcStream() {
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	lis, err := net.Listen(conf.ListenNet, "127.0.0.1:")
	StreamRpcListenPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	logger.INFO("GameAgent lis: ", conf.ListenNet, " addr: ", StreamRpcListenPort)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRouteConnectGameServer(grpcServer, &StreamServer{""})
	logger.INFO("GameApp started!")
	go grpcServer.Serve(lis)
}

// Per stream for per goroutine
func (s *StreamServer) AgentStream(stream pb.RouteConnectGame_AgentStreamServer) error {
	return s.startReceiver(stream)
}

func (s *StreamServer) DeployScene(ctx context.Context, in *pb.DeploySceneRequest) (*pb.DeploySceneReply, error) {
	success := scene_mgr.TryLoadScene(in.GetSceneId())
	return &pb.DeploySceneReply{Success: success}, nil
}

func (s *StreamServer) RequestPlayer(ctx context.Context, in *pb.RequestPlayerRequest) (*pb.RequestPlayerReply, error) {
	rpcRsp, err := HandleRPCCall(in.GetAccountId(), in.GetData())
	if err != nil {
		return nil, err
	}
	return &pb.RequestPlayerReply{
		Data: rpcRsp,
	}, nil
}

func (s *StreamServer) startReceiver(stream pb.RouteConnectGame_AgentStreamServer) error {
	logger.INFO("gameAgent startReceiver")
	headers, _ := metadata.FromIncomingContext(stream.Context())
	accountId := headers["accountId"][0]
	accountConnectMap.Store(accountId, stream)
	PlayerConnected(accountId, stream)
	atomic.AddInt32(&onlinePlayers, 1)

	// Receiving client msg
	var err error
	var in *pb.RouteMsg
	for {
		in, err = stream.Recv()
		if err == io.EOF {
			logger.ERR("GameAgent EOF")
			accountConnectMap.Delete(accountId)
			break
		}
		if err != nil {
			logger.ERR("GameAgent err: ", err)
			break
		}
		logger.INFO("AgentStream received: ", accountId, " data: ", len(in.GetData()))
		HandleRequest(accountId, in.GetData())
	}

	accountConnectMap.Delete(accountId)
	PlayerDisconnected(accountId)
	atomic.AddInt32(&onlinePlayers, -1)

	return err
}

