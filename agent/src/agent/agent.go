package main

import (
	"net"
	"goslib/logger"
	"goslib/redisdb"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"connection"
	"time"
	"context"
	"goslib/utils"
	"strconv"
)

func main() {
	// Start redis
	conf := gosconf.REDIS_FOR_SERVICE
	redisdb.Connect(conf.Host, conf.Password, conf.Db)

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

	connectAgentMgr(l.Addr().(*net.TCPAddr).Port)

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

func connectAgentMgr(listenPort int) {
	OutIP, _ := utils.GetOutboundIP()
	LocalIP, _ := utils.GetLocalIp()
	logger.INFO("OutboundIP: ", OutIP, " LocalIP: ", LocalIP)
	conf := gosconf.RPC_FOR_CONNECT_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		logger.ERR("connect AgentMgr failed: ", err)
		return
	}
	client := pb.NewDispatcherClient(conn)

	go func() {
		// Report Agent Info
		var host string
		var err error
		if gosconf.IS_DEBUG {
			host = "127.0.0.1"
		} else {
			host, err = utils.GetPublicIP()
		}
		if err != nil {
			time.Sleep(gosconf.HEARTBEAT)
			go connectAgentMgr(listenPort)
			return
		}

		port := strconv.Itoa(listenPort)
		uuid := utils.GenId([]string{host, port})
		agentInfo := &pb.AgentInfo{
			Uuid: uuid,
			Host: host,
			Port: port,
			Ccu: 0,
		}

		// Heartbeat
		for {
			ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
			defer cancel()
			agentInfo.Ccu = connection.OnlinePlayers()
			_, err = client.ReportAgentInfo(ctx, agentInfo)
			if err != nil {
				logger.ERR("ReportAgentInfo heartbeat failed: ", err)
			}
			logger.INFO("ReportAgentInfo: ", agentInfo.Uuid, " Host: ", host, " Port: ", port, " ccu: ", agentInfo.Ccu)
			time.Sleep(gosconf.HEARTBEAT)
		}
	}()
}
