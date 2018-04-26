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
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisdb.Connect(conf.Host, conf.Password, conf.Db)

	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()
	leaderboard.Start()
	gslib.Run()
	scene_mgr.StartSceneMgr()
	player.StartPlayerManager()
	player.StartRpcStream()
}

