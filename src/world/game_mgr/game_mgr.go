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
package game_mgr

import (
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

var rpcServer *grpc.Server

type gameAppMgr struct {
}

func Start() {
	go startGameAppMgrRPC()
}

func Stop() {
	rpcServer.GracefulStop()
}

func startGameAppMgrRPC() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	lis, err := net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	if err != nil {
		logger.ERR("failed to listen: ", err)
		panic(err)
	}
	logger.INFO("GameAppMgr lis: ", conf.ListenNet, " port: ", conf.ListenPort)

	rpcServer = grpc.NewServer()
	proto.RegisterGameDispatcherServer(rpcServer, &gameAppMgr{})
	reflection.Register(rpcServer)
	if err := rpcServer.Serve(lis); err != nil {
		logger.ERR("failed to serve: ", err)
	}
}

// DispatchPlayer implements connectAppProto.DispatchPlayer
func (s *gameAppMgr) DispatchGame(ctx context.Context, in *proto.DispatchGameRequest) (*proto.DispatchGameReply, error) {
	info, err := DispatchGame(in.AccountId, in.SceneId)
	if err != nil {
		return nil, err
	}

	return &proto.DispatchGameReply{
		GameAppId:   info.AppId,
		GameAppHost: info.AppHost,
		GameAppPort: info.AppPort,
	}, err
}
