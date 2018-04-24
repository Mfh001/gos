package main

import (
	"ConnectAppMgr"
	"GameAppMgr"
	"goslib/redisDB"
	"gosconf"
)

func main() {
	conf := gosconf.REDIS_FOR_SERVICE
	redisDB.Connect(conf.Host, conf.Password, conf.Db)
	ConnectAppMgr.Start()
	GameAppMgr.Start()
}
