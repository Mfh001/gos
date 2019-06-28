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
	"github.com/gorilla/websocket"
	"github.com/mafei198/gos/goslib/game_server/connection"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/rs/xid"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
)

type WSAgent struct {
	mt      int
	uuid    string
	wsConn  *websocket.Conn
	connIns *connection.Connection
}

var websocketListener net.Listener

func StartWSAgent() {
	http.HandleFunc("/", wsHandler)

	tcpConf := gosconf.TCP_SERVER_GAME

	var err error
	if gosconf.START_TYPE == gosconf.START_TYPE_K8S {
		websocketListener, err = net.Listen("tcp", net.JoinHostPort("", tcpConf.ListenPort))
	} else {
		websocketListener, err = net.Listen("tcp", ":0")
	}
	if err != nil {
		logger.ERR("Connection listen failed: ", err)
		panic(err)
	}
	AgentPort = strconv.Itoa(websocketListener.Addr().(*net.TCPAddr).Port)
	logger.INFO("WSAgent lis: ", AgentPort)
	go func() {
		if err := http.Serve(websocketListener, nil); err != nil {
			logger.ERR("start WSAgent failed: ", err)
			panic(err)
		}
	}()
}

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if !enableAcceptConn {
		return
	}

	logger.INFO("WSAgent accepted new conn")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ERR("upgrade:", err)
		return
	}
	defer c.Close()

	agent := &WSAgent{}
	connIns := connection.New(agent)
	agent.uuid = xid.New().String()
	agent.connIns = connIns
	agent.wsConn = c

	defer func() {
		connIns.Cleanup()
	}()

	atomic.AddInt32(&OnlinePlayers, 1)
	for {
		mt, message, err := c.ReadMessage()
		agent.mt = mt
		if enableAcceptMsg {
			if err != nil {
				logger.ERR("read: ", err)
				break
			}
			if err = agent.OnMessage(message); err != nil {
				break
			}
		}
	}
	atomic.AddInt32(&OnlinePlayers, -1)
	agent.OnDisconnected()
}

func (self *WSAgent) GetUuid() string {
	return self.uuid
}

func (self *WSAgent) OnMessage(data []byte) error {
	logger.INFO("ws received: ", data)
	err := self.connIns.OnMessage(data)
	return err
}

func (self *WSAgent) SendMessage(data []byte) error {
	err := self.wsConn.WriteMessage(self.mt, data)
	if err != nil {
		logger.ERR("write: ", err)
		return err
	}
	return nil
}

func (self *WSAgent) OnDisconnected() {
	self.connIns.OnDisconnected(self.uuid)
}

func (self *WSAgent) ConnectScene(session *session_utils.Session) error {
	return self.connIns.ConnectScene(session)
}
