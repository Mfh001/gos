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
	return castChannel(channel, "Join", playerId, handler)
}

func Leave(channel, playerId string) error {
	return castChannel(channel, "Leave", playerId)
}

func Publish(channel, playerId, category string, data interface{}) error {
	msg := &BroadcastMsg{
		Category: category,
		Channel:  channel,
		SenderId: playerId,
		Data:     data,
	}
	return castChannel(channel, "Publish", msg)
}

func castChannel(channel string, args ...interface{}) error {
	if !gen_server.Exists(channel) {
		err := StartChannel(channel)
		if err != nil {
			logger.ERR("start channel failed: ", err)
			return err
		}
	}
	gen_server.Cast(channel, args...)
	return nil
}

/*
   GenServer Callbacks
*/
func (self *Broadcast) Init(args []interface{}) (err error) {
	self.subscribers = make(map[string]MsgHandler)
	return nil
}

func (self *Broadcast) HandleCast(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "Join" {
		self.handleJoin(args[1].(string), args[1].(MsgHandler))
	} else if method_name == "Leave" {
		self.handleLeave(args[1].(string))
	} else if method_name == "Publish" {
		self.handlePublish(args[1].(*BroadcastMsg))
	}
}

func (self *Broadcast) HandleCall(args []interface{}) (interface{}, error) {
	return nil, nil
}

func (self *Broadcast) Terminate(reason string) (err error) {
	self.subscribers = nil
	return nil
}

/*
   Callback Handlers
*/

func (self *Broadcast) handleJoin(playerId string, handler MsgHandler) {
	self.subscribers[playerId] = handler
}

func (self *Broadcast) handleLeave(playerId string) {
	if _, ok := self.subscribers[playerId]; ok {
		delete(self.subscribers, playerId)
	}
}

func (self *Broadcast) handlePublish(msg *BroadcastMsg) {
	for _, handler := range self.subscribers {
		handler(msg)
	}
}
