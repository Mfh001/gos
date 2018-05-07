package main

import (
	"app/register"
	"app/register/callbacks"
	"app/register/tables"
	"context"
	"google.golang.org/grpc"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/broadcast"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/memstore"
	"goslib/redisdb"
	"goslib/utils"
	"gslib"
	"gslib/leaderboard"
	"gslib/player"
	"gslib/scene_mgr"
	"time"
)

func main() {
	redisdb.InitServiceClient()

	// Register routes
	register.Load()
	// Register MySQL data loader
	register.RegisterDataLoader()
	// Register cutom MySQL data loader
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()
	leaderboard.Start()

	// Print routine info
	go gslib.SysRoutine()
	// Init DB Connections
	memstore.InitDB()
	memstore.StartDBPersister()
	tables.RegisterTables(memstore.GetSharedDBInstance())

	// Start broadcast server
	gen_server.Start(gslib.BROADCAST_SERVER_ID, new(broadcast.Broadcast))

	// Start scene manager
	scene_mgr.StartSceneMgr()

	player.StartPlayerRPC()
	player.StartPlayerManager()

	// Start listen agent stream
	player.StartRpcStream()

	// connect to game manager, retry if failed
	for {
		if err := connectGameMgr(); err != nil {
			logger.ERR("connect GameMgr failed: ", err)
		}
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func connectGameMgr() error {
	OutIP, _ := utils.GetOutboundIP()
	LocalIP, _ := utils.GetLocalIp()
	logger.INFO("OutboundIP: ", OutIP, " LocalIP: ", LocalIP)
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(conf.DialAddress, conf.DialOptions...)
	if err != nil {
		return err
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
		return err
	}

	uuid := utils.GenId([]string{host, player.StreamRpcListenPort})
	player.CurrentGameAppId = uuid
	gameInfo := &pb.ReportGameRequest{
		Uuid: uuid,
		Host: host,
		Port: player.StreamRpcListenPort,
	}

	for {
		heartbeat(client, gameInfo)
		time.Sleep(gosconf.HEARTBEAT)
	}

	return nil
}

func heartbeat(client pb.GameDispatcherClient, gameInfo *pb.ReportGameRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()
	gameInfo.Ccu = player.OnlinePlayers()
	_, err := client.ReportGameInfo(ctx, gameInfo)
	if err != nil {
		logger.ERR("ReportGameInfo heartbeat failed: ", err)
	}
	logger.INFO("ReportGameInfo: ", gameInfo.Uuid, " Host: ", gameInfo.Host, " Port: ", gameInfo.Port, " ccu: ", gameInfo.Ccu)
}
