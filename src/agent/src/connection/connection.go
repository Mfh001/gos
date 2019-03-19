package connection

import (
	"errors"
	"gen/api/pt"
	"gen/proto"
	"goslib/api"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/packet"
	"goslib/session_utils"
	"sync"
	"sync/atomic"
)

type AgentBehavior interface {
	OnMessage(data []byte) error
	SendMessage(data []byte) error
}

type Connection struct {
	id        int64
	authed    bool
	agent     AgentBehavior
	processed int64
	session   *session_utils.Session
	streams   map[int]proto.RouteConnectGame_AgentStreamClient
}

var connectionMap = &sync.Map{}
var onlinePlayers int32
var connId int64

func OnlinePlayers() int32 {
	return onlinePlayers
}

func New(agent AgentBehavior) *Connection {
	instance := &Connection{
		id:        connId,
		authed:    false,
		processed: 0,
		agent:     agent,
		streams:   make(map[int]proto.RouteConnectGame_AgentStreamClient),
	}
	connectionMap.Store(instance.id, instance)
	atomic.AddInt64(&connId, 1)
	atomic.AddInt32(&onlinePlayers, 1)
	return instance
}

func (self *Connection) Cleanup() {
	defer func() {
		connectionMap.Delete(self.id)
		self.agent = nil
		self.session = nil
		self.authed = false
		atomic.AddInt32(&onlinePlayers, -1)
		for _, stream := range self.streams {
			err := stream.CloseSend()
			if err != nil {
				logger.ERR("Connection close stream failed: ", err)
			}
		}
		self.streams = nil
	}()
}

func (self *Connection) OnMessage(data []byte) error {
	var err error
	if self.authed {
		if err = self.proxyRequest(data); err != nil {
			logger.ERR("SendMsg to GameApp failed: ", self.session.AccountId, " err: ", err)
			return err
		}
	} else {
		if self.authed, err = self.authConn(data); err != nil {
			logger.ERR("AuthConn failed: ", err)
			return err
		}
		if !self.authed {
			logger.ERR("AuthConn failed! ")
			return errors.New("AuthConn failed")
		}
	}
	return nil
}

func (self *Connection) setupProxy(protocolType int) (proto.RouteConnectGame_AgentStreamClient, error) {
	var game *game_utils.Game
	var err error

	switch protocolType {
	case pt.PT_TYPE_GS:
		if self.session.GameAppId != "" {
			game, err = game_utils.Find(self.session.GameAppId)
			if err != nil {
				return nil, err
			}
		} else {
			game, err = ChooseGameServer(self.session)
			if err != nil {
				return nil, err
			}
		}
	case pt.PT_TYPE_ROOM:
		if self.session.RoomAppId != "" {
			game, err = game_utils.Find(self.session.RoomAppId)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("room app id is blank")
		}
	}

	stream, err := ConnectGameServer(game.Uuid, self.session.AccountId, self.session.RoomId, protocolType, self.agent)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return nil, err
	}
	return stream, nil
}

func (self *Connection) proxyRequest(data []byte) error {
	stream, err := self.getStream(data)
	if err != nil {
		logger.ERR("Get Stream failed: ", err)
		return err
	}

	err = stream.Send(&proto.RouteMsg{
		Data: data,
	})
	return err
}

func (self *Connection) getStream(data []byte) (proto.RouteConnectGame_AgentStreamClient, error) {
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	protocolType := pt.IdToType[protocol]

	stream, ok := self.streams[protocolType]
	if !ok {
		stream, err := self.setupProxy(protocolType)
		if err != nil {
			return nil, err
		}
		self.streams[protocolType] = stream
	}
	return stream, nil
}

func (self *Connection) authConn(data []byte) (bool, error) {
	// Decode data
	params, err := decodeAuthData(data)
	if err != nil {
		return false, err
	}

	// Validate Token from AuthApp
	success, err := self.validateSession(params)
	if err != nil {
		return false, err
	}

	// Send auth response to client
	writer, err := api.Encode("SessionAuthResponse", &pt.SessionAuthResponse{Success: success})
	if err != nil {
		return false, err
	}

	self.processed++
	// INFO("Processed: ", self.processed, " Response Data: ", response_data)
	if self.agent != nil {
		err = self.agent.SendMessage(writer.GetSendData())
	}

	return success, err
}

func decodeAuthData(data []byte) (*pt.SessionAuthParams, error) {
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := pt.IdToName[protocol]

	if decode_method != "SessionAuthParams" {
		return nil, errors.New("Request UnAuthed connection: " + decode_method)
	}

	params, err := api.Decode(decode_method, reader)
	if err != nil {
		return nil, err
	}
	return params.(*pt.SessionAuthParams), err
}

func (self *Connection) validateSession(params *pt.SessionAuthParams) (bool, error) {
	session, err := session_utils.Find(params.AccountId)
	if err != nil {
		return false, err
	}

	if session.Token != params.Token {
		return false, errors.New("session token invalid")
	}

	self.session = session

	return true, nil
}
