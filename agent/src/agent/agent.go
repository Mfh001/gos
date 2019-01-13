package main

import (
	"connection"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/logger"
	"net"
)

func main() {
	// Start redis
	connectGameMgr()
	connection.StartProxyManager()

	// Listen incomming tcp connections
	tcpConf := gosconf.TCP_SERVER_CONNECT_APP
	l, err := net.Listen(tcpConf.Network, tcpConf.Address)
	if err != nil {
		logger.ERR("Connection listen failed: ", err)
	}
	logger.INFO("ConnectApp started!")
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.ERR("Connection accept failed: ", err)
		}
		connection.Handle(conn)
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
