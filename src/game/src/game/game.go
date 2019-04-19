package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"app/gen_register"
	"gosconf"
	"goslib/game_server"
	"goslib/redisdb"
	"os"
)

func main() {
	var role string
	if len(os.Args) > 1 {
		role = os.Args[1]
	} else {
		role = gosconf.GS_ROLE_DEFAULT
	}
	redisdb.StartClient()
	println("game server start with role: ", role)
	game_server.Start(role, func() {
		gen_register.RegisterRoutes()
		custom_register.RegisterCustomDataLoader()
		callbacks.RegisterSceneLoad()
	})
}
