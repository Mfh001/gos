package main

import (
	"agent_mgr"
	"game_mgr"
	"goslib/redisdb"
	//"RootMgr"
)

func main() {
	redisdb.InitServiceClient()
	go agent_mgr.Start()
	game_mgr.Start()
	//RootMgr.Start()
}
