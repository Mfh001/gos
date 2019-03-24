package broadcast

import (
	"goslib/gen_server"
)

const SERVER = "__broadcast_mgr_server__"

type BroadcastMgr struct {
}

func StartMgr() {
	gen_server.Start(SERVER, new(BroadcastMgr))
}

func StartChannel(channel string) error {
	_, err := gen_server.Call(SERVER, "StartChannel", channel)
	return err
}

func (self *BroadcastMgr) Init(args []interface{}) (err error) {
	return nil
}

func (self *BroadcastMgr) HandleCast(args []interface{}) {
}

func (self *BroadcastMgr) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "StartChannel" {
		channel := args[1].(string)
		if !gen_server.Exists(channel) {
			gen_server.Start(channel, new(Broadcast))
		}
	}
	return nil, nil
}

func (self *BroadcastMgr) Terminate(reason string) (err error) {
	return nil
}
