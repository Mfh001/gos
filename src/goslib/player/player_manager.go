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
package player

import (
	"errors"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/session_utils"
	"time"
)

const SERVER = "__player_manager_server__"

type Status int
type PlayerStatus int

const (
	working Status = iota
	shutdown
)

const (
	MAX_SLEEP   = 30 // seconds
	GC_INTERVAL = 10 * time.Second
)

/*
   GenServer Callbacks
*/
type PlayerManager struct {
	status          Status
	players         map[string]*gen_server.GenServer
	sleepingPlayers map[string]int64
	sleepingGCTimer *time.Ticker
	shutdowning     map[string]bool
}

func StartManager() error {
	_, err := gen_server.Start(SERVER, new(PlayerManager))
	return err
}

func DelShutdown(playerId string) {
	gen_server.Cast(SERVER, &DelShutDownParams{playerId})
}

func EnsureShutdown() {
	gen_server.Cast(SERVER, &ShutdownPlayersParams{})
	for {
		result, err := gen_server.Call(SERVER, &RemainPlayersParams{})
		logger.INFO("EnsureShutdown: ", result, err)
		if err == nil && result.(int) == 0 {
			return
		}
		if err != nil {
			logger.ERR("get remainPlayers failed: ", err)
		}
		logger.INFO("Stopping players remain: ", result.(int))
		time.Sleep(1 * time.Second)
	}
}

func StartPlayer(accountId string) error {
	var sceneId string
	if !IsTesting {
		session, err := session_utils.Find(accountId)
		if err != nil {
			logger.ERR("StartPlayer failed: ", err)
			return err
		}
		if session.GameAppId != CurrentGameAppId {
			err = errors.New("player not belongs to this server")
			logger.ERR("StartPlayer failed: ", err)
			return err
		}
		sceneId = session.SceneId
	}
	_, err := gen_server.Call(SERVER, &StartPlayerParams{
		accountId: accountId,
		sceneId:   sceneId,
	})
	return err
}

func SleepPlayer(accountId string) {
	gen_server.Cast(SERVER, &SleepPlayerParams{accountId})
}

var ticker = &TickerGCParams{}

func (self *PlayerManager) Init(args []interface{}) (err error) {
	self.players = make(map[string]*gen_server.GenServer)
	self.sleepingPlayers = make(map[string]int64)
	self.sleepingGCTimer = time.NewTicker(GC_INTERVAL)
	self.shutdowning = map[string]bool{}
	go func() {
		var err error
		for range self.sleepingGCTimer.C {
			_, err = gen_server.Call(SERVER, ticker)
			if err != nil {
				logger.ERR("player_manager tickerPersist failed: ", err)
			}
		}
	}()
	return nil
}

func (self *PlayerManager) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *DelShutDownParams:
		self.handleDelShutdown(params)
		break
	case *SleepPlayerParams:
		self.handleSleepPlayer(params.accountId)
		break
	case *ShutdownPlayersParams:
		self.handleShutdownPlayers()
		break
	default:
		logger.ERR("player_manager unhandle message: ", params)
	}
}

func (self *PlayerManager) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *StartPlayerParams:
		return self.handleStartPlayer(params)
	case *RemainPlayersParams:
		return self.handleRemainPlayers(), nil
	case *TickerGCParams:
		self.tickerGC()
		break
	default:
		logger.ERR("player_manager unhandle message: ", params)
	}
	return nil, nil
}

func (self *PlayerManager) Terminate(reason string) (err error) {
	self.sleepingGCTimer.Stop()
	return nil
}

type StartPlayerParams struct {
	accountId string
	sceneId   string
}

func (self *PlayerManager) handleStartPlayer(params *StartPlayerParams) (interface{}, error) {
	if self.status != working {
		return nil, errors.New("player_manager is shutting down")
	}
	// wakeup sleeping player
	if _, ok := self.sleepingPlayers[params.accountId]; ok {
		server := self.players[params.accountId]
		server.Cast(&ActiveSleepParams{})
		gen_server.SetGenServer(params.accountId, server)
		return nil, nil
	}
	// start player
	if !gen_server.Exists(params.accountId) {
		server, err := gen_server.Start(params.accountId, new(Player), params.accountId, params.sceneId)
		if err != nil {
			return nil, err
		}
		self.players[params.accountId] = server
	}
	return nil, nil
}

type SleepPlayerParams struct{ accountId string }

func (self *PlayerManager) handleSleepPlayer(accountId string) {
	gen_server.DelGenServer(accountId)
	self.sleepingPlayers[accountId] = time.Now().Unix()
}

type ShutdownPlayersParams struct{}

func (self *PlayerManager) handleShutdownPlayers() {
	self.sleepingGCTimer.Stop()
	self.status = shutdown
	self.batchShutdownPlayers()
}

type RemainPlayersParams struct{}

func (self *PlayerManager) handleRemainPlayers() int {
	return len(self.players)
}

type DelShutDownParams struct{ playerId string }

func (self *PlayerManager) handleDelShutdown(params *DelShutDownParams) {
	delete(self.players, params.playerId)
	delete(self.sleepingPlayers, params.playerId)
}

func (self *PlayerManager) batchShutdownPlayers() {
	notSleepped := 0
	for accountId, server := range self.players {
		if _, ok := self.sleepingPlayers[accountId]; !ok {
			self.handleSleepPlayer(accountId)
			notSleepped++
		} else {
			self.shutdownPlayer(accountId, server)
		}
	}

	if notSleepped > 0 {
		logger.WARN("retry batchShutdown, notSleeped: ", notSleepped)
		gen_server.Cast(SERVER, &ShutdownPlayersParams{})
		return
	}

	for accountId, server := range self.players {
		self.shutdownPlayer(accountId, server)
	}
}

type TickerGCParams struct{}

func (self *PlayerManager) tickerGC() {
	now := time.Now().Unix()
	for accountId, sleepAt := range self.sleepingPlayers {
		if now-sleepAt > MAX_SLEEP {
			if server, ok := self.players[accountId]; ok {
				if err := server.Stop("Shutdown inActive player!"); err != nil {
					logger.ERR("Shutdown inActive player failed: ", err)
				} else {
					delete(self.players, accountId)
					delete(self.sleepingPlayers, accountId)
				}
			} else {
				delete(self.sleepingPlayers, accountId)
			}
		}
	}
}

func (self *PlayerManager) shutdownPlayer(playerId string, playerServer *gen_server.GenServer) {
	if _, ok := self.shutdowning[playerId]; ok {
		return
	}
	self.shutdowning[playerId] = true
	go retry(func() error {
		if err := playerServer.Stop("shutdown"); err != nil {
			logger.ERR("player shutdown failed: ", playerId, err)
			return err
		} else {
			logger.INFO("player shutdown success: ", playerId)
			DelShutdown(playerId)
			return nil
		}
	})
}

const Backoff = 1
const BackoffRatio = 2
const MaxBackoff = 3

func retry(handler func() error) {
	backoff := Backoff
	for i := 0; ; i++ {
		if err := handler(); err == nil {
			return
		}
		time.Sleep(time.Duration(backoff) * time.Second)
		if i < MaxBackoff {
			backoff *= BackoffRatio
		}
	}
}
