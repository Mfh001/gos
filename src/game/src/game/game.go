package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"gen/register"
	"gen/register/tables"
	"gosconf"
	"goslib/broadcast"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/memstore"
	"goslib/player"
	"goslib/player_rpc"
	"goslib/redisdb"
	"goslib/scene_mgr"
	"goslib/timertask"
	"goslib/utils"
	"time"
)

func main() {
	// Register routes
	register.Load()
	// Register MySQL data loader
	register.RegisterDataLoader()
	// Register cutom MySQL data loader
	custom_register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()
	//leaderboard.Start()

	// Print routine info
	go utils.SysRoutine()
	// Init DB Connections
	memstore.InitDB()
	memstore.StartDBPersister()
	tables.RegisterTables(memstore.GetSharedDBInstance())

	// Start broadcast server
	broadcast.Start()

	// Start scene manager
	scene_mgr.StartSceneMgr()

	player_rpc.StartPlayerRPC()
	player.StartPlayerManager()

	// Start listen agent stream
	StartRpcStream()

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

	uuid := utils.GenId([]string{host, StreamRpcListenPort})
	player.CurrentGameAppId = uuid

	app, err := addGame(uuid, host, StreamRpcListenPort)
	if err != nil {
		logger.ERR("addGame failed: ", err)
		return err
	}
	logger.INFO("AddGame: ", uuid, " Host: ", host, " Port: ", StreamRpcListenPort)

	for {
		heartbeat(app)
		time.Sleep(gosconf.HEARTBEAT)
	}

	return nil
}

func heartbeat(app *game_utils.Game) {
	// TODO for k8s health check
	app.Ccu = OnlinePlayers()
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
	_, err := redisdb.Instance().SAdd(gosconf.RK_GAME_APP_IDS, app.Uuid).Result()
	if err != nil {
		return nil, err
	}
	err = app.Save()
	return app, err
}
