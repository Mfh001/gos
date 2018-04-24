package player

import (
	"goslib/gen_server"
)

const SERVER = "__player_manager_server__"

/*
   GenServer Callbacks
*/
type PlayerManager struct {
}

func StartPlayerManager() {
	gen_server.Start(SERVER, new(PlayerManager))
}

func StartPlayer(accountId string) {
	gen_server.Call(SERVER, "StartPlayer", accountId)
}

func (self *PlayerManager) Init(args []interface{}) (err error) {
	return nil
}

func (self *PlayerManager) HandleCast(args []interface{}) {
}

func (self *PlayerManager) HandleCall(args []interface{}) interface{} {
	handle := args[0].(string)
	if handle == "StartPlayer" {
		accountId := args[1].(string)
		if !gen_server.Exists(accountId) {
			gen_server.Start(accountId, new(Player))
		}
	}
	return nil
}

func (self *PlayerManager) Terminate(reason string) (err error) {
	return nil
}