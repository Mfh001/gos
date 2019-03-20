package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"gosconf"
	"goslib/game_server"
)

func main() {
	game_server.Start(gosconf.GAME_DOMAIN, func() {
		custom_register.RegisterCustomDataLoader()
		callbacks.RegisterBroadcast()
		callbacks.RegisterSceneLoad()
	})
}
