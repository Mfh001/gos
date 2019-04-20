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
	_, err := gen_server.Call(SERVER, &StartChannelParams{channel})
	return err
}

func (self *BroadcastMgr) Init(args []interface{}) (err error) {
	return nil
}

func (self *BroadcastMgr) HandleCast(req *gen_server.Request) {
}

type StartChannelParams struct {
	channel string
}

func (self *BroadcastMgr) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *StartChannelParams:
		if !gen_server.Exists(params.channel) {
			gen_server.Start(params.channel, new(Broadcast))
		}
		break
	}
	return nil, nil
}

func (self *BroadcastMgr) Terminate(reason string) (err error) {
	return nil
}
