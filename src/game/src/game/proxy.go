package main

import (
	"context"
	"gen/api/pt"
	"gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gosconf"
	"goslib/logger"
	"goslib/player"
	"goslib/scene_mgr"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
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
	lis, err := net.Listen(conf.ListenNet, conf.ListenAddr)
	StreamRpcListenPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	logger.INFO("GameAgent lis: ", conf.ListenNet, " addr: ", StreamRpcListenPort)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterRouteConnectGameServer(grpcServer, &StreamServer{""})
	logger.INFO("GameApp started!")
	go grpcServer.Serve(lis)
}

// Per stream for per goroutine
func (s *StreamServer) AgentStream(stream proto.RouteConnectGame_AgentStreamServer) error {
	return s.startReceiver(stream)
}

func (s *StreamServer) DeployScene(ctx context.Context, in *proto.DeploySceneRequest) (*proto.DeploySceneReply, error) {
	success := scene_mgr.TryLoadScene(in.GetSceneId())
	return &proto.DeploySceneReply{Success: success}, nil
}

func (s *StreamServer) RequestPlayer(ctx context.Context, in *proto.RequestPlayerRequest) (*proto.RequestPlayerReply, error) {
	rpcRsp, err := player.HandleRPCCall(in.GetAccountId(), in.GetData())
	if err != nil {
		return nil, err
	}
	return &proto.RequestPlayerReply{
		Data: rpcRsp,
	}, nil
}

func (s *StreamServer) startReceiver(stream proto.RouteConnectGame_AgentStreamServer) error {
	logger.INFO("gameAgent startReceiver")
	headers, _ := metadata.FromIncomingContext(stream.Context())
	logger.INFO(headers)

	protocolType, err := strconv.Atoi(headers["protocoltype"][0])
	if err != nil {
		logger.ERR("parse protocolType failed: ", err)
		return err
	}
	accountId := headers["accountid"][0]
	roomId := headers["roomid"][0]

	var actorId string
	switch protocolType {
	case pt.PT_TYPE_GS:
		actorId = accountId
		break
	case pt.PT_TYPE_ROOM:
		actorId = roomId
		break
	}

	accountConnectMap.Store(actorId, stream)
	player.Connected(actorId, stream)
	atomic.AddInt32(&onlinePlayers, 1)

	// Receiving client msg
	var in *proto.RouteMsg
	for {
		in, err = stream.Recv()
		if err == io.EOF {
			logger.ERR("GameAgent EOF")
			accountConnectMap.Delete(actorId)
			break
		}
		if err != nil {
			logger.ERR("GameAgent err: ", err)
			break
		}
		logger.INFO("AgentStream received: ", actorId, " data: ", len(in.GetData()))
		player.HandleRequest(actorId, in.GetData())
	}

	accountConnectMap.Delete(actorId)
	player.Disconnected(actorId)
	atomic.AddInt32(&onlinePlayers, -1)

	return err
}
