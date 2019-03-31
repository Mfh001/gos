package agent

import (
	"gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gosconf"
	"goslib/game_server/connection"
	"goslib/logger"
	"net"
	"sync/atomic"
)

type StreamAgent struct {
	connIns *connection.Connection
	stream  proto.GameStreamAgent_GameStreamServer
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
	proto.RegisterGameStreamAgentServer(grpcServer, &StreamAgent{})
	logger.INFO("GameApp started!")
	go grpcServer.Serve(streamListener)
}

// Per stream for per goroutine
func (s *StreamAgent) GameStream(stream proto.GameStreamAgent_GameStreamServer) error {
	headers, _ := metadata.FromIncomingContext(stream.Context())
	accountId := headers["accountid"][0]
	roomAppId := headers["roomappid"][0]
	roomId := headers["roomid"][0]

	agent := &StreamAgent{}
	connIns := connection.New(agent)

	defer func() {
		connIns.Cleanup()
	}()

	agent.connIns = connIns
	agent.stream = stream
	s.connIns.RoomPlayerConnected(accountId, roomAppId, roomId)

	var err error
	var in *proto.StreamAgentMsg
	atomic.AddInt32(&OnlinePlayers, 1)
	for {
		in, err = stream.Recv()
		if !enableAcceptMsg {
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
	self.connIns.OnDisconnected()
}
