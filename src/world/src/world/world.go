package main

import (
	"auth"
	"cache_mgr"
	"game_mgr"
	"goslib/redisdb"
)

func main() {
	redisdb.StartClient()
	auth.Start()
	cache_mgr.Start()
	game_mgr.Start()
}
