package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"app/gen_register"
	"goslib/game_server"
)

func main() {
	game_server.Start(func() {
		gen_register.RegisterRoutes()
		custom_register.RegisterCustomDataLoader()
		callbacks.RegisterSceneLoad()
	})
}
