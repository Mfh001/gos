package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"gen/register"
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
	go utils.SysRoutine()

	register.RegisterRoutes()
	register.RegisterDataLoader()
	register.RegisterTables(memstore.GetSharedDBInstance())

	custom_register.RegisterCustomDataLoader()

	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()

	err := memstore.StartDB()
	if err != nil {
		panic(err.Error())
	}
	memstore.StartDBPersister()

	broadcast.Start()

	scene_mgr.Start()

	player.StartManager()
	player_rpc.Start()

	timertask.Start()

	StartRpcStream()

	host, err := gameHost()
	if err != nil {
		logger.ERR("game get host failed: ", err)
		return
	}

	uuid := gameUuid(host)

	player.CurrentGameAppId = uuid

	for {
		if err := reportGameInfo(uuid, host); err != nil {
			logger.ERR("reportGameInfo failed: ", err)
		}
		time.Sleep(gosconf.HEARTBEAT)
	}
}

func reportGameInfo(uuid, host string) error {
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

func gameHost() (string, error) {
	hostname, err := utils.GetHostname()
	if err != nil {
		return "", err
	}
	host := hostname + "." + gosconf.GAME_DOMAIN
	return host, nil
}

func gameUuid(host string) string {
	uuid := utils.GenId([]string{host, StreamRpcListenPort})
	return uuid
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

func heartbeat(app *game_utils.Game) {
	// TODO for k8s health check
	app.Ccu = OnlinePlayers()
	app.ActiveAt = time.Now().Unix()
	err := app.Save()

	if err != nil {
		logger.ERR("game heartbeat failed: ", err)
	} else {
		logger.INFO("ReportGameInfo: ", app.Uuid, " Host: ", app.Host, " Port: ", app.Port, " ccu: ", app.Ccu)
	}
}
