package main

import (
	"app/register"
	"app/register/callbacks"
	"gslib"
	"gslib/leaderboard"
	"gosconf"
	"goslib/redisdb"
	"gslib/player"
	"gslib/scene_mgr"
	"time"
	"goslib/logger"
	"google.golang.org/grpc"
	"goslib/utils"
	pb "gos_rpc_proto"
	"context"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisdb.Connect(conf.Host, conf.Password, conf.Db)

	// Register routes
	register.Load()
	// Register MySQL data loader
	register.RegisterDataLoader()
	// Register cutom MySQL data loader
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()
	leaderboard.Start()
	gslib.Run()
	// Start scene manager
	scene_mgr.StartSceneMgr()
	// Start Player goroutine manager
	player.StartPlayerManager()
	// Start listen agent stream
	player.StartRpcStream()
	// connect to Game manager
	connectGameMgr()
}

func connectGameMgr() {
	OutIP, _ := utils.GetOutboundIP()
	LocalIP, _ := utils.GetLocalIp()
	logger.INFO("OutboundIP: ", OutIP, " LocalIP: ", LocalIP)
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		logger.ERR("connect AgentMgr failed: ", err)
		return
	}
	client := pb.NewGameDispatcherClient(conn)

	// Report Game Info
	var host string
	if gosconf.IS_DEBUG {
		host = "127.0.0.1"
	} else {
		host, err = utils.GetLocalIp()
	}
	if err != nil {
		time.Sleep(gosconf.HEARTBEAT)
		connectGameMgr()
		return
	}

	uuid := utils.GenId([]string{host, player.StreamRpcListenPort})
	gameInfo := &pb.ReportGameRequest{
		Uuid: uuid,
		Host: host,
		Port: player.StreamRpcListenPort,
	}

	// Heartbeat
	for {
		ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
		defer cancel()
		gameInfo.Ccu = player.OnlinePlayers()
		_, err = client.ReportGameInfo(ctx, gameInfo)
		if err != nil {
			logger.ERR("ReportGameInfo heartbeat failed: ", err)
		}
		logger.INFO("ReportGameInfo: ", gameInfo.Uuid, " Host: ", host, " Port: ", gameInfo.Port, " ccu: ", gameInfo.Ccu)
		time.Sleep(gosconf.HEARTBEAT)
	}
}
