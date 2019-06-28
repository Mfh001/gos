/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package game_server

import (
	"context"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/scene_mgr"
	"google.golang.org/grpc"
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
	rpcRsp, err := player.HandleRPCCall(in.GetAccountId(), in.GetCategory(), in.GetData())
	if err != nil {
		return nil, err
	}
	return &proto.RequestPlayerReply{
		Data: rpcRsp,
	}, nil
}
