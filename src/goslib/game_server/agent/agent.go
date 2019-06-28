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
	"google.golang.org/grpc"
	"net"
)

var OnlinePlayers int32
var enableAcceptConn bool
var enableAcceptMsg bool

var AgentPort string

func Start() {
	connectGameMgr()
	connection.StartProxyManager()

	enableAcceptConn = true
	enableAcceptMsg = true

	switch gosconf.AGENT_PROTOCOL {
	case gosconf.AGENT_PROTOCOL_TCP:
		StartTCPAgent()
		break
	case gosconf.AGENT_PROTOCOL_WS:
		StartWSAgent()
		break
	}

	StartStreamAgent()
}

func StopAcceptor() {
	enableAcceptConn = false
	switch gosconf.AGENT_PROTOCOL {
	case gosconf.AGENT_PROTOCOL_TCP:
	case gosconf.AGENT_PROTOCOL_WS:
		if err := websocketListener.Close(); err != nil {
			logger.ERR("Close weboscket listener failed: ", err)
		}
	}
	if err := streamListener.Close(); err != nil {
		logger.ERR("Close stream listener failed: ", err)
	}
}

func StopAcceptMsg() {
	enableAcceptMsg = false
}

func connectGameMgr() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(net.JoinHostPort(gosconf.GetWorldIP(), conf.ListenPort), conf.DialOptions...)
	if err != nil {
		logger.ERR("connection connectGameMgr failed: ", err)
		return
	}

	connection.GameMgrRpcClient = proto.NewGameDispatcherClient(conn)
}
