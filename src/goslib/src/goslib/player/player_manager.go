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
	gen_server.Cast(SERVER, &DelShutDownParams{playerId})
}

func ShutdownPlayers() error {
	_, err := gen_server.Call(SERVER, &ShutdownPlayersParams{})
	return err
}

func EnsureShutdown() {
	err := ShutdownPlayers()
	if err != nil {
		logger.ERR("ShutdownPlayers failed: ", err)
	}
	for {
		result, err := gen_server.Call(SERVER, &RemainPlayersParams{})
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
	_, err = gen_server.Call(SERVER, &StartPlayerParams{accountId})
	return err
}

func (self *PlayerManager) Init(args []interface{}) (err error) {
	self.players = make(map[string]bool)
	return nil
}

func (self *PlayerManager) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *DelShutDownParams:
		self.handleDelShutdown(params)
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
	case *ShutdownPlayersParams:
		self.handleShutdownPlayers()
		break
	default:
		logger.ERR("player_manager unhandle message: ", params)
	}
	return nil, nil
}

func (self *PlayerManager) Terminate(reason string) (err error) {
	return nil
}

type StartPlayerParams struct { accountId string }
func (self *PlayerManager) handleStartPlayer(params *StartPlayerParams) (interface{}, error) {
	if self.status != working {
		return nil, errors.New("player_manager is shutting down")
	}
	if !gen_server.Exists(params.accountId) {
		_, err := gen_server.Start(params.accountId, new(Player), params.accountId)
		if err != nil {
			return nil, err
		}
		self.players[params.accountId] = true
	}
	return nil, nil
}

type ShutdownPlayersParams struct {}
func (self *PlayerManager) handleShutdownPlayers() {
	self.status = shutdown
	self.batchShutdownPlayers()
}

type RemainPlayersParams struct {}
func (self *PlayerManager) handleRemainPlayers() int {
	return len(self.players)
}

type DelShutDownParams struct { playerId string }
func (self *PlayerManager) handleDelShutdown(params *DelShutDownParams) {
	delete(self.players, params.playerId)
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
