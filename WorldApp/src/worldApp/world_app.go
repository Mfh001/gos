package main

import (
	"ConnectAppMgr"
	"GameAppMgr"
	"goslib/redisDB"
)

func main() {
	redisDB.Connect("localhost:6379", "", 0)
	ConnectAppMgr.Start()
	GameAppMgr.Start()
}
