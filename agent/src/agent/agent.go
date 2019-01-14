package main

import (
	"connection"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/logger"
)

func main() {
	connectGameMgr()
	connection.StartProxyManager()

	switch gosconf.AGENT_PROTOCOL {
	case gosconf.AGENT_PROTOCOL_TCP:
		StartWSAgent()
		break
	case gosconf.AGENT_PROTOCOL_WS:
		StartTCPAgent()
		break
	}
}

func connectGameMgr() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		logger.ERR("connection connectGameMgr failed: ", err)
		return
	}

	connection.GameMgrRpcClient = pb.NewGameDispatcherClient(conn)
}

func heartbeat() {
	// TODO for k8s health check
}
