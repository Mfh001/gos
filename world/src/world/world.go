package main

import (
	"agent_mgr"
	"game_mgr"
	"goslib/redisdb"
	"gosconf"
	//"RootMgr"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisdb.Connect(conf.Host, conf.Password, conf.Db)
	go agent_mgr.Start()
	game_mgr.Start()
	//RootMgr.Start()
}
