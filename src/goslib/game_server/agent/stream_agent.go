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

package agent

import (
	"github.com/mafei198/gos/goslib/game_server/connection"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"sync/atomic"
)

type StreamAgentServer struct {
}

var streamListener net.Listener

func StartStreamAgent() {
	var err error
	conf := gosconf.RPC_FOR_GAME_APP_STREAM
	streamListener, err = net.Listen(conf.ListenNet, net.JoinHostPort("", conf.ListenPort))
	logger.INFO("StreamAgent lis: ", conf.ListenNet, " port: ", conf.ListenPort)
	if err != nil {
		logger.ERR("failed to listen: ", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGameStreamAgentServer(grpcServer, &StreamAgentServer{})
	logger.INFO("GameApp started!")
	go grpcServer.Serve(streamListener)
}

// Per stream for per goroutine
func (s *StreamAgentServer) GameStream(stream proto.GameStreamAgent_GameStreamServer) error {
	headers, _ := metadata.FromIncomingContext(stream.Context())
	accountId := headers["accountid"][0]
	sceneId := headers["sceneid"][0]
	return startAgent(accountId, sceneId, stream)
}

func startAgent(accountId, sceneId string, stream proto.GameStreamAgent_GameStreamServer) error {
	agent := &StreamAgent{}
	connIns := connection.New(agent)

	defer func() {
		connIns.Cleanup()
	}()

	agent.uuid = xid.New().String()
	agent.connIns = connIns
	agent.stream = stream
	agent.connIns.ScenePlayerConnected(accountId, sceneId)

	var err error
	var in *proto.StreamAgentMsg
	atomic.AddInt32(&OnlinePlayers, 1)
	for {
		in, err = stream.Recv()
		if enableAcceptMsg {
			if err != nil {
				logger.ERR("GameAgent err: ", err)
				break
			}
			if err = agent.OnMessage(in.GetData()); err != nil {
				break
			}
		}
	}
	atomic.AddInt32(&OnlinePlayers, -1)
	agent.OnDisconnected()
	return err
}

type StreamAgent struct {
	uuid    string
	connIns *connection.Connection
	stream  proto.GameStreamAgent_GameStreamServer
}

func (self *StreamAgent) GetUuid() string {
	return self.uuid
}

func (self *StreamAgent) OnMessage(data []byte) error {
	logger.INFO("stream received: ", data)
	err := self.connIns.OnMessage(data)
	return err
}

func (self *StreamAgent) SendMessage(data []byte) error {
	err := self.stream.Send(&proto.StreamAgentMsg{
		Data: data,
	})
	if err != nil {
		logger.ERR("write: ", err)
		return err
	}
	return nil
}

func (self *StreamAgent) OnDisconnected() {
	self.connIns.OnDisconnected(self.uuid)
}

func (self *StreamAgent) ConnectScene(session *session_utils.Session) error {
	return self.connIns.ConnectScene(session)
}
