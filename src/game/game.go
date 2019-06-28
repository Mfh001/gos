/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"github.com/mafei198/gos/game/api_server"
	"github.com/mafei198/gos/game/app/custom_register/callbacks"
	"github.com/mafei198/gos/game/app/gen_register"
	"github.com/mafei198/gos/game/app/life_cycles"
	"github.com/mafei198/gos/goslib/database"
	"github.com/mafei198/gos/goslib/game_server"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/redisdb"
	"os"
)

func main() {
	life_cycles.BeforeGameStart()

	redisdb.StartClient()
	database.StartMongo()

	var role string
	if len(os.Args) > 1 {
		role = os.Args[1]
	} else {
		role = gosconf.GS_ROLE_DEFAULT
	}
	println("game server start with role: ", role)
	customRegister := func() {
		gen_register.RegisterRoutes()
		callbacks.RegisterSceneLoad()
	}

	api_server.StartApiServer()

	game_server.Start(role, customRegister, life_cycles.AfterGameStart, life_cycles.AfterGameShutdown)
}
