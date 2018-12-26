package main

import (
	"app/register"
	"app/register/callbacks"
	"app/register/tables"
	"gosconf"
	"goslib/broadcast"
	"goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/memstore"
	"goslib/redisdb"
	"goslib/utils"
	"gslib"
	"gslib/player"
	"gslib/player_rpc"
	"gslib/scene_mgr"
	"gslib/timertask"
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
	//leaderboard.Start()

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

	player_rpc.StartPlayerRPC()
	player.StartPlayerManager()

	// Start listen agent stream
	player.StartRpcStream()

	timertask.Start()

	// connect to game manager, retry if failed
	for {
		if err := reportGameInfo(); err != nil {
			logger.ERR("connect GameMgr failed: ", err)
		}
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func reportGameInfo() error {
	// Report Game Info
	hostname, err := utils.GetHostname()
	host := hostname + "." + gosconf.GAME_DOMAIN
	if err != nil {
		return err
	}

	uuid := utils.GenId([]string{host, player.StreamRpcListenPort})
	player.CurrentGameAppId = uuid

	app, err := addGame(uuid, host, player.StreamRpcListenPort)
	if err != nil {
		logger.ERR("addGame failed: ", err)
		return err
	}
	logger.INFO("AddGame: ", uuid, " Host: ", host, " Port: ", player.StreamRpcListenPort)

	for {
		heartbeat(app)
		time.Sleep(gosconf.HEARTBEAT)
	}

	return nil
}

func heartbeat(app *game_utils.Game) {
	// TODO for k8s health check
	app.Ccu = player.OnlinePlayers()
	app.ActiveAt = time.Now().Unix()
	app.Save()
	logger.INFO("ReportGameInfo: ", app.Uuid, " Host: ", app.Host, " Port: ", app.Port, " ccu: ", app.Ccu)
}

func addGame(uuid, host, port string) (*game_utils.Game, error) {
	app := &game_utils.Game{
		Uuid:     uuid,
		Host:     host,
		Port:     port,
		ActiveAt: time.Now().Unix(),
	}
	_, err := redisdb.ServiceInstance().SAdd(gosconf.RK_GAME_APP_IDS, app.Uuid).Result()
	if err != nil {
		return nil, err
	}
	err = app.Save()
	return app, err
}
