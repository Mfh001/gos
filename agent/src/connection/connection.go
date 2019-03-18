package connection

import (
	"api"
	"api/pt"
	"errors"
	"github.com/json-iterator/go"
	pb "gos_rpc_proto"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/packet"
	"goslib/session_utils"
	"sync"
	"sync/atomic"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
	stream    pb.RouteConnectGame_AgentStreamClient
}

var sessionMap = &sync.Map{}
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
	}
	connectionMap.Store(instance.id, instance)
	atomic.AddInt64(&connId, 1)
	atomic.AddInt32(&onlinePlayers, 1)
	return instance
}

func (self *Connection) Cleanup() {
	defer func() {
		connectionMap.Delete(self.id)
		self.stream = nil
		self.agent = nil
		self.session = nil
		self.authed = false
		atomic.AddInt32(&onlinePlayers, -1)
		if self.stream != nil {
			err := self.stream.CloseSend()
			if err != nil {
				logger.ERR("Connection close stream failed: ", err)
			}
		}
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
		if err = self.setupProxy(); err != nil {
			logger.ERR("DispatchToGameApp failed: ", err)
			return err
		}
	}
	return nil
}

func (self *Connection) setupProxy() error {
	sessionMap.Store(self.session.AccountId, self.session)
	var game *game_utils.Game
	var err error
	if self.session.GameAppId != "" {
		game, err = game_utils.Find(self.session.GameAppId)
		if err != nil {
			return err
		}
	} else {
		game, err = ChooseGameServer(self.session)
		if err != nil {
			return err
		}
	}
	self.stream, err = ConnectGameServer(game.Uuid, self.session.AccountId, self.agent)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return err
	}
	return nil
}

func (self *Connection) proxyRequest(data []byte) error {
	if self.stream == nil {
		return errors.New("Stream not exits!")
	}
	err := self.stream.Send(&pb.RouteMsg{
		Data: data,
	})
	return err
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
		self.agent.SendMessage(writer.GetSendData())
	}

	return success, nil
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
		return false, errors.New("Session token invalid!")
	}

	self.session = session

	return true, nil
}
