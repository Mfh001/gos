package main

import (
	"net"
	"goslib/logger"
	"connection"
	"goslib/redisDB"
	"google.golang.org/grpc"
	pb "gosRpcProto"
	"gosconf"
)

/*
 * 连接服务
 *
 * 连接校验
 * 消息转发
 * 广播管理
 */

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisDB.Connect(conf.Host, conf.Password, conf.Db)
	connectGameMgr()
	connection.Setup()

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
		connection.Start(conn)
	}
}

func connectGameMgr() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		logger.ERR("connection connectGameMgr failed: ", err)
	}

	connection.GameMgrRpcClient = pb.NewGameDispatcherClient(conn)
}

