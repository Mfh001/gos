package connection

import (
	"errors"
	"gen/api/pt"
	"gen/proto"
	"goslib/api"
	"goslib/game_server/interfaces"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/packet"
	"goslib/player"
	"goslib/session_utils"
	"sync"
	"sync/atomic"
)

const (
	CATEGORY_PLAYER = iota
	CATEGORY_ROOM
)

type Connection struct {
	id        int64
	category  int
	accountId string
	authed    bool
	agent     interfaces.AgentBehavior
	processed int64
	session   *session_utils.Session
	streams   map[int]proto.GameStreamAgent_GameStreamClient

	roomAppId string
	roomId    string
}

var connectionMap = &sync.Map{}
var onlinePlayers int32
var connId int64

func OnlinePlayers() int32 {
	return onlinePlayers
}

func New(agent interfaces.AgentBehavior) *Connection {
	instance := &Connection{
		id:        connId,
		category:  CATEGORY_PLAYER,
		authed:    false,
		processed: 0,
		agent:     agent,
		streams:   make(map[int]proto.GameStreamAgent_GameStreamClient),
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
		self.streams = nil
	}()
}

func (self *Connection) OnMessage(data []byte) error {
	switch self.category {
	case CATEGORY_PLAYER:
		return self.OnPlayerMessage(data)
	case CATEGORY_ROOM:
		return self.OnRoomMessage(data)
	}
	return nil
}

func (self *Connection) OnPlayerMessage(data []byte) error {
	var err error
	if self.authed {
		protocolType, err := api.ParseProtolType(data)
		if err != nil {
			return err
		}
		switch protocolType {
		case pt.PT_TYPE_GS:
			return player.HandleRequest(self.accountId, self.accountId, data)
		case pt.PT_TYPE_ROOM:
			stream, err := self.getStream(protocolType, data)
			if err != nil {
				logger.ERR("getStream failed: ", err)
				return err
			}
			return stream.Send(&proto.StreamAgentMsg{
				Data: data,
			})
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
		return player.Connected(self.accountId, self.accountId, self.agent)
	}
	return nil
}

func (self *Connection) OnRoomMessage(data []byte) error {
	if !self.authed {
		session, err := session_utils.Find(self.accountId)
		if err != nil {
			return err
		}
		if session.RoomAppId == self.roomAppId && session.RoomId == self.roomId {
			self.authed = true
			player.Connected(self.roomId, self.accountId, self.agent)
		} else {
			return errors.New("AuthConn failed")
		}
	}
	return player.HandleRequest(self.roomId, self.accountId, data)
}

func (self *Connection) OnDisconnected() {
	if self.accountId != "" {
		switch self.category {
		case CATEGORY_PLAYER:
			_ = player.Disconnected(self.accountId, self.accountId)
		case CATEGORY_ROOM:
			_ = player.Disconnected(self.roomId, self.accountId)
		}
	}
}

func (self *Connection) RoomPlayerConnected(accountId, roomAppId, roomId string) {
	self.category = CATEGORY_ROOM
	self.accountId = accountId
	self.roomAppId = roomAppId
	self.roomId = roomId
}

func (self *Connection) setupProxy() (proto.GameStreamAgent_GameStreamClient, error) {
	var game *game_utils.Game
	var err error

	if self.session.RoomAppId != "" {
		game, err = game_utils.Find(self.session.RoomAppId)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("room app id is blank")
	}
	stream, err := ConnectGameServer(game.Uuid, self.session.AccountId, self.session.RoomAppId, self.session.RoomId, self.agent)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return nil, err
	}
	return stream, nil
}

func (self *Connection) getStream(protocolType int, data []byte) (proto.GameStreamAgent_GameStreamClient, error) {
	stream, ok := self.streams[protocolType]
	if !ok {
		stream, err := self.setupProxy()
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
		data, err := writer.GetSendData()
		if err != nil {
			return false, err
		}
		err = self.agent.SendMessage(data)
	}

	self.accountId = params.AccountId

	return success, err
}

func decodeAuthData(data []byte) (*pt.SessionAuthParams, error) {
	reader := packet.Reader(data)
	protocol, err := reader.ReadUint16()
	if err != nil {
		return nil, err
	}
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
		logger.INFO("sessionToken: ", session.Token, " paramsToken: ", params.Token, " accountId: ", params.AccountId)
		return false, errors.New("session token invalid")
	}

	self.session = session

	return true, nil
}
