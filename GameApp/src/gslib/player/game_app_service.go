package player

import (
	"gosconf"
	"net"
	"io"
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"goslib/logger"
	"context"
	"sync"
	"gslib/sceneMgr"
	"google.golang.org/grpc/metadata"
)

type StreamServer struct {
	ConnectAppId string
}

var accountConnectMap *sync.Map

func StartRpcStream() {
	accountConnectMap = &sync.Map{}
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	logger.INFO("GameAgent lis: ", conf.ListenNet, " addr: ", conf.ListenAddr)
	lis, err := net.Listen(conf.ListenNet, conf.ListenAddr)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRouteConnectGameServer(grpcServer, &StreamServer{""})
	logger.INFO("GameApp started!")
	grpcServer.Serve(lis)
}

// Per stream for per goroutine
func (s *StreamServer) AgentStream(stream pb.RouteConnectGame_AgentStreamServer) error {
	return s.startReceiver(stream)
}

func (s *StreamServer) DeployScene(ctx context.Context, in *pb.DeploySceneRequest) (*pb.DeploySceneReply, error) {
	success := sceneMgr.TryLoadScene(in.GetSceneId())
	return &pb.DeploySceneReply{Success: success}, nil
}

func (s *StreamServer) startReceiver(stream pb.RouteConnectGame_AgentStreamServer) error {
	logger.INFO("gameAgent startReceiver")
	headers, _ := metadata.FromIncomingContext(stream.Context())
	accountId := headers["session"][0]
	accountConnectMap.Store(accountId, stream)

	PlayerConnected(accountId, stream)

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
		logger.INFO("AgentStream received: ", in.GetAccountId(), " data: ", len(in.GetData()))
		HandleRequest(accountId, in.GetData())
	}

	accountConnectMap.Delete(accountId)
	PlayerDisconnected(accountId)

	return err
}
