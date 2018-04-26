package main

import (
	"app/register"
	"app/register/callbacks"
	"gslib"
	"gslib/leaderboard"
	"gosconf"
	"goslib/redisDB"
	"gslib/player"
	"gslib/sceneMgr"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisDB.Connect(conf.Host, conf.Password, conf.Db)

	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	callbacks.RegisterSceneLoad()
	leaderboard.Start()
	gslib.Run()
	sceneMgr.StartSceneMgr()
	player.StartPlayerManager()
	player.StartRpcStream()
}

