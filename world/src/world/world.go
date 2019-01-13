package main

import (
	"agent_mgr"
	"game_mgr"
	//"RootMgr"
)

func main() {
	go agent_mgr.Start()
	game_mgr.Start()
	//RootMgr.Start()
}
