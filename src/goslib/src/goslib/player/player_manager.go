package player

import (
	"errors"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/session_utils"
	"time"
)

const SERVER = "__player_manager_server__"

type Status int
type PlayerStatus int

const (
	working Status = iota
	shutdown
)

/*
   GenServer Callbacks
*/
type PlayerManager struct {
	status  Status
	players map[string]bool
}

func StartManager() error {
	_, err := gen_server.Start(SERVER, new(PlayerManager))
	return err
}

func DelShutdown(playerId string) {
	gen_server.Cast(SERVER, "delShutdown", playerId)
}

func ShutdownPlayers() error {
	_, err := gen_server.Call(SERVER, "shutdownPlayers")
	return err
}

func EnsureShutdown() {
	err := ShutdownPlayers()
	if err != nil {
		logger.ERR("ShutdownPlayers failed: ", err)
	}
	for {
		result, err := gen_server.Call(SERVER, "remainPlayers")
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
	_, err = gen_server.Call(SERVER, "StartPlayer", accountId)
	return err
}

func (self *PlayerManager) Init(args []interface{}) (err error) {
	self.players = make(map[string]bool)
	return nil
}

func (self *PlayerManager) HandleCast(args []interface{}) {
	handle := args[0].(string)
	switch handle {
	case "delShutdown":
		playerId := args[1].(string)
		delete(self.players, playerId)
		break
	default:
		logger.ERR("player_manager unhandle message: ", handle)
	}
}

func (self *PlayerManager) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	switch handle {
	case "StartPlayer":
		if self.status != working {
			return nil, errors.New("player_manager is shutting down")
		}
		accountId := args[1].(string)
		if !gen_server.Exists(accountId) {
			_, err := gen_server.Start(accountId, new(Player), accountId)
			if err != nil {
				return nil, err
			}
			self.players[accountId] = true
		}
		break
	case "remainPlayers":
		return len(self.players), nil
	case "shutdownPlayers":
		self.status = shutdown
		self.batchShutdownPlayers()
		break
	default:
		logger.ERR("player_manager unhandle message: ", handle)
	}
	return nil, nil
}

func (self *PlayerManager) Terminate(reason string) (err error) {
	return nil
}

func (self *PlayerManager) batchShutdownPlayers() {
	for accountId := range self.players {
		go retry(func() error {
			if err := gen_server.Stop(accountId, "shutdown"); err != nil {
				logger.ERR("player_manager shutdown player failed: ", accountId, err)
				return err
			} else {
				logger.INFO("player_manager shutdown player success: ", accountId)
				DelShutdown(accountId)
				return nil
			}
		})
	}
}

const Backoff = 1
const BackoffRatio = 2
const RetryTimes = 5

func retry(handler func() error) {
	backoff := Backoff
	for i := 0; i < RetryTimes; i++ {
		if err := handler(); err == nil {
			return
		}
		time.Sleep(time.Second)
		backoff *= BackoffRatio
	}
}
