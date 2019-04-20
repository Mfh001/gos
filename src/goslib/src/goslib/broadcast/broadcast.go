package broadcast

import (
	"goslib/gen_server"
	"goslib/logger"
)

type BroadcastMsg struct {
	Category string
	Channel  string
	SenderId string
	Data     interface{}
}

type MsgHandler func(msg *BroadcastMsg)

type Broadcast struct {
	subscribers map[string]MsgHandler
}

func Join(channel, playerId string, handler MsgHandler) error {
	return castChannel(channel, &JoinParams{playerId, handler})
}

func Leave(channel, playerId string) error {
	return castChannel(channel, &LeaveParams{playerId})
}

func Publish(channel, playerId, category string, data interface{}) error {
	msg := &BroadcastMsg{
		Category: category,
		Channel:  channel,
		SenderId: playerId,
		Data:     data,
	}
	return castChannel(channel, &PublishParams{msg})
}

func castChannel(channel string, msg interface{}) error {
	if !gen_server.Exists(channel) {
		err := StartChannel(channel)
		if err != nil {
			logger.ERR("start channel failed: ", err)
			return err
		}
	}
	gen_server.Cast(channel, msg)
	return nil
}

/*
   GenServer Callbacks
*/
func (self *Broadcast) Init(args []interface{}) (err error) {
	self.subscribers = make(map[string]MsgHandler)
	return nil
}

func (self *Broadcast) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *JoinParams:
		self.handleJoin(params)
		break
	case *LeaveParams:
		self.handleLeave(params)
		break
	case *PublishParams:
		self.handlePublish(params)
		break
	}
}

func (self *Broadcast) HandleCall(req *gen_server.Request) (interface{}, error) {
	return nil, nil
}

func (self *Broadcast) Terminate(reason string) (err error) {
	self.subscribers = nil
	return nil
}

/*
   Callback Handlers
*/

type JoinParams struct {
	playerId string
	handler MsgHandler
}
func (self *Broadcast) handleJoin(params *JoinParams) {
	self.subscribers[params.playerId] = params.handler
}

type LeaveParams struct { playerId string }
func (self *Broadcast) handleLeave(params *LeaveParams) {
	if _, ok := self.subscribers[params.playerId]; ok {
		delete(self.subscribers, params.playerId)
	}
}

type PublishParams struct {msg *BroadcastMsg}
func (self *Broadcast) handlePublish(params *PublishParams) {
	for _, handler := range self.subscribers {
		handler(params.msg)
	}
}
