package main

import (
	"connection"
	"context"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/logger"
	"goslib/redisdb"
	"goslib/utils"
	"net"
	"strconv"
	"time"
)

func main() {
	// Start redis
	redisdb.InitServiceClient()
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

	go func() {
		for {
			if err := connectAgentMgr(l.Addr().(*net.TCPAddr).Port); err != nil {
				logger.ERR("connect AgentMgr failed: ", err)
			}
			time.Sleep(gosconf.HEARTBEAT)
		}
	}()

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

func connectAgentMgr(listenPort int) error {
	OutIP, _ := utils.GetOutboundIP()
	LocalIP, _ := utils.GetLocalIp()
	logger.INFO("OutboundIP: ", OutIP, " LocalIP: ", LocalIP)
	conf := gosconf.RPC_FOR_CONNECT_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	defer conn.Close()

	if err != nil {
		return err
	}
	client := pb.NewDispatcherClient(conn)

	// Report Agent Info
	var host string
	if gosconf.IS_DEBUG {
		host = "127.0.0.1"
	} else {
		if host, err = utils.GetPublicIP(); err != nil {
			return err
		}
	}

	port := strconv.Itoa(listenPort)
	uuid := utils.GenId([]string{host, port})
	agentInfo := &pb.AgentInfo{
		Uuid: uuid,
		Host: host,
		Port: port,
		Ccu:  0,
	}

	// heartbeat
	for {
		heartbeat(client, agentInfo)
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func heartbeat(client pb.DispatcherClient, agentInfo *pb.AgentInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	agentInfo.Ccu = connection.OnlinePlayers()
	_, err := client.ReportAgentInfo(ctx, agentInfo)
	if err != nil {
		logger.ERR("ReportAgentInfo heartbeat failed: ", err)
	}
	logger.INFO("ReportAgentInfo: ", agentInfo.Uuid, " Host: ", agentInfo.Host, " Port: ", agentInfo.Port, " ccu: ", agentInfo.Ccu)
}
