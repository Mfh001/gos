package main

import (
	"ConnectAppMgr"
	"GameAppMgr"
	"goslib/redisDB"
	"gosconf"
	//"RootMgr"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisDB.Connect(conf.Host, conf.Password, conf.Db)
	go ConnectAppMgr.Start()
	GameAppMgr.Start()
	//RootMgr.Start()
}
