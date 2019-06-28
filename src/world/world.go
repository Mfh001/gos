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
	"github.com/mafei198/gos/goslib/database"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/world/api_mgr"
	"github.com/mafei198/gos/world/auth"
	"github.com/mafei198/gos/world/cache_mgr"
	"github.com/mafei198/gos/world/game_mgr"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	redisdb.StartClient()
	database.StartMongo()

	session_utils.Start()

	auth.Start()
	cache_mgr.Start()
	game_mgr.Start()
	api_mgr.Start()

	<-stopChan // wait for SIGINT or SIGTERM
	logger.INFO("Shutting world service...")

	shutdown()

	logger.INFO("world service stopped")
}

func shutdown() {
	// stop auth service
	auth.Stop()

	// stop game mgr
	game_mgr.Stop()

	// stop cache mgr
	cache_mgr.Stop()
}
