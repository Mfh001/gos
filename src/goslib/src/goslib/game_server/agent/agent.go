package agent

import (
	"gen/proto"
	"google.golang.org/grpc"
	"gosconf"
	"goslib/game_server/connection"
	"goslib/logger"
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
