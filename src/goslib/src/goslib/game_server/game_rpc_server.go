package game_server

import (
	"context"
	"gen/proto"
	"google.golang.org/grpc"
	"gosconf"
	"goslib/logger"
	"goslib/player"
	"goslib/scene_mgr"
	"net"
)

type StreamServer struct {
	ConnectAppId string
}

func StartRpcStream() {
	conf := gosconf.RPC_FOR_GAME_APP_RPC
	lis, err := net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	logger.INFO("GameRpcServer lis: ", conf.ListenNet, " port: ", conf.ListenPort)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGameRpcServerServer(grpcServer, &StreamServer{""})
	logger.INFO("GameApp started!")
	go grpcServer.Serve(lis)
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
