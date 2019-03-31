package main

import (
	"auth"
	"cache_mgr"
	"game_mgr"
	"goslib/logger"
	"goslib/redisdb"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	redisdb.StartClient()

	auth.Start()
	cache_mgr.Start()
	game_mgr.Start()

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
