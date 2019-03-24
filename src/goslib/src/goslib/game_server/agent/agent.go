package agent

import (
	"gen/proto"
	"google.golang.org/grpc"
	"gosconf"
	"goslib/game_server/connection"
	"goslib/logger"
)

var OnlinePlayers int32

func Start() {
	connectGameMgr()
	connection.StartProxyManager()

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

func connectGameMgr() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		logger.ERR("connection connectGameMgr failed: ", err)
		return
	}

	connection.GameMgrRpcClient = proto.NewGameDispatcherClient(conn)
}
