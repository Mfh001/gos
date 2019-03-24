package main

import (
	"auth"
	"game_mgr"
)

func main() {
	auth.Start()
	game_mgr.Start()
}
