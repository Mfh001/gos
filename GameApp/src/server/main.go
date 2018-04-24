package main

import (
	"app/register"
	"app/register/callbacks"
	"gslib"
	"gslib/leaderboard"
	"gosconf"
	"goslib/redisDB"
	"gslib/player"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisDB.Connect(conf.Host, conf.Password, conf.Db)

	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	leaderboard.Start()
	gslib.Run()
	player.StartRpcStream()
}

