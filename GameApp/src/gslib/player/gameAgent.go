package player

import (
	"gosconf"
	"net"
	"log"
	"io"
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"goslib/logger"
	"goslib/sessionMgr"
	"context"
	"sync"
)

type StreamServer struct {
	Authed bool
	ConnectAppId string
}

var accountConnectMap *sync.Map

func StartRpcStream() {
	accountConnectMap = &sync.Map{}
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	lis, err := net.Listen(conf.ListenNet, conf.ListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRouteConnectGameServer(grpcServer, &StreamServer{false, ""})
	grpcServer.Serve(lis)

}

func GetConnectId(accountId string) string {
	v, ok := accountConnectMap.Load(accountId)
	if ok {
		return v.(string)
	} else {
		return ""
	}
}

// Per stream for per goroutine
func (s *StreamServer) AgentStream(stream pb.RouteConnectGame_AgentStreamServer) error {
	return s.startReceiver(stream)
}

func (s *StreamServer) AgentRegister(ctx context.Context, in *pb.RegisterMsg) (*pb.RegisterReply, error) {
	if in.GetIsRegister() {
		accountConnectMap.Store(in.GetAccountId(), in.GetConnectAppId())
	} else {
		accountConnectMap.Delete(in.GetAccountId())
	}
	return &pb.RegisterReply{Status: true}, nil
}

func (s *StreamServer) startReceiver(stream pb.RouteConnectGame_AgentStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			logger.ERR("")
			return nil
		}
		if err != nil {
			return err
		}

		logger.INFO("AgentStream received: ", in.GetAccountId(), " data: ", len(in.GetData()))

		if !s.Authed {
			session, err := sessionMgr.Find(in.GetAccountId())
			if err != nil {
				logger.ERR("AgentStream Auth failed: ", in.GetAccountId(), " err: ", err)
				continue
			}
			s.Authed = true
			s.ConnectAppId = session.ConnectAppId
			StartGameAgentSender(s.ConnectAppId, stream)
		} else {
			HandleRequest(in.GetAccountId(), in.GetData())
		}
	}
}
