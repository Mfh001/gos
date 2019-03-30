package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"app/gen_register"
	"goslib/game_server"
	"goslib/redisdb"
)

func main() {
	redisdb.StartClient()
	game_server.Start(func() {
		gen_register.RegisterRoutes()
		custom_register.RegisterCustomDataLoader()
		callbacks.RegisterSceneLoad()
	})
}
