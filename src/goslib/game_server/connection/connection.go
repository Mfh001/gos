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

package connection

import (
	"errors"
	"github.com/mafei198/gos/goslib/api"
	"github.com/mafei198/gos/goslib/game_server/interfaces"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/packet"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/scene_utils"
	"github.com/mafei198/gos/goslib/session_utils"
	"sync"
	"sync/atomic"
)

const (
	CATEGORY_PLAYER = iota
	CATEGORY_SCENE
)

type Connection struct {
	id        int64
	category  int
	accountId string
	authed    bool
	agent     interfaces.AgentBehavior
	processed int64
	session   *session_utils.Session
	stream    proto.GameStreamAgent_GameStreamClient

	sceneId string
}

var connectionMap = &sync.Map{}
var onlinePlayers int32
var connId int64

func New(agent interfaces.AgentBehavior) *Connection {
	atomic.AddInt64(&connId, 1)
	atomic.AddInt32(&onlinePlayers, 1)
	instance := &Connection{
		id:        connId,
		category:  CATEGORY_PLAYER,
		authed:    false,
		processed: 0,
		agent:     agent,
	}
	connectionMap.Store(instance.id, instance)
	return instance
}

func (self *Connection) Cleanup() {
	defer func() {
		connectionMap.Delete(self.id)
		self.agent = nil
		self.session = nil
		self.authed = false
		atomic.AddInt32(&onlinePlayers, -1)
		self.stream = nil
	}()
}

func (self *Connection) OnMessage(data []byte) error {
	switch self.category {
	case CATEGORY_PLAYER:
		return self.OnPlayerMessage(data)
	case CATEGORY_SCENE:
		return self.OnSceneMessage(data)
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
			switch gosconf.SCENE_CONNECT_MODE {
			case gosconf.SCENE_CONNECT_MODE_DIRECT:
				return self.OnSceneMessage(data)
			case gosconf.SCENE_CONNECT_MODE_PROXY:
				stream, err := self.getStream()
				if err != nil {
					logger.ERR("getStream failed: ", err)
					return err
				}
				return stream.Send(&proto.StreamAgentMsg{
					Data: data,
				})
			}
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

func (self *Connection) OnSceneMessage(data []byte) error {
	if !self.authed {
		if self.session != nil && self.session.SceneId == self.sceneId {
			self.authed = true
			err := player.Connected(self.sceneId, self.accountId, self.agent)
			if err != nil {
				return err
			}
		} else {
			return errors.New("AuthConn failed")
		}
	}
	return player.HandleRequest(self.sceneId, self.accountId, data)
}

func (self *Connection) OnDisconnected(connId string) {
	if self.accountId != "" {
		switch self.category {
		case CATEGORY_PLAYER:
			player.Disconnected(self.accountId, self.accountId, connId)
		case CATEGORY_SCENE:
			player.Disconnected(self.sceneId, self.accountId, connId)
		}
	}
}

func (self *Connection) ScenePlayerConnected(accountId, sceneId string) {
	self.category = CATEGORY_SCENE
	self.accountId = accountId
	self.sceneId = sceneId
	self.session, _ = session_utils.Find(self.accountId)
}

func (self *Connection) getStream() (proto.GameStreamAgent_GameStreamClient, error) {
	if self.stream != nil {
		return self.stream, nil
	}
	return self.stream, self.ConnectScene(self.session)
}

func (self *Connection) ConnectScene(session *session_utils.Session) error {
	scene, err := scene_utils.Find(session.SceneId)
	if err != nil {
		return err
	}
	stream, err := ConnectGameServer(scene.GameAppId, session.AccountId, session.SceneId, self.agent)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return err
	}
	self.stream = stream
	return nil
}

func (self *Connection) authConn(data []byte) (bool, error) {
	// Decode data
	reqId, params, err := decodeAuthData(data)
	if err != nil {
		return false, err
	}

	// Validate Token from AuthApp
	success, err := self.validateSession(params)
	if err != nil {
		return false, err
	}

	// Send auth response to client
	writer, err := api.Encode(&pt.SessionAuthResponse{Success: success})
	if err != nil {
		return false, err
	}

	self.processed++
	// INFO("Processed: ", self.processed, " Response Data: ", response_data)
	if self.agent != nil {
		data, err := writer.GetSendData(reqId)
		if err != nil {
			return false, err
		}
		err = self.agent.SendMessage(data)
	}

	self.accountId = params.AccountId

	return success, err
}

func decodeAuthData(data []byte) (reqId int32, params *pt.SessionAuthParams, err error) {
	reader := packet.Reader(data)
	reqId, err = reader.ReadInt32()
	if err != nil {
		return
	}
	protocol, err := reader.ReadUint16()
	if err != nil {
		return
	}
	decode_method := pt.IdToName[protocol]

	if decode_method != pt.PT_SessionAuthParams {
		err = errors.New("Request UnAuthed connection: " + decode_method)
		return
	}

	paramsInterface, err := api.Decode(decode_method, reader)
	if err != nil {
		return
	}
	params = paramsInterface.(*pt.SessionAuthParams)
	return
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
