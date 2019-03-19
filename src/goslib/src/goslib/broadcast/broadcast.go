package broadcast

import (
	"goslib/gen_server"
)

const BROADCAST_SERVER_ID = "__broadcast_server__"

type BroadcastMsg struct {
	Category string
	Channel  string
	SenderId string
	Data     interface{}
}

type Broadcast struct {
	channels map[string](map[string]bool)
}

func Start() {
	gen_server.Start(BROADCAST_SERVER_ID, new(Broadcast))
}

func JoinChannel(playerId, channel string) {
	gen_server.Cast(BROADCAST_SERVER_ID, "JoinChannel", playerId, channel)
}

func LeaveChannel(playerId, channel string) {
	gen_server.Cast(BROADCAST_SERVER_ID, "LeaveChannel", playerId, channel)
}

func PublishChannelMsg(playerId, channel, category string, data interface{}) {
	msg := &BroadcastMsg{
		Category: category,
		Channel:  channel,
		SenderId: playerId,
		Data:     data,
	}
	gen_server.Cast(BROADCAST_SERVER_ID, "Publish", msg)
}

/*
   GenServer Callbacks
*/
func (self *Broadcast) Init(args []interface{}) (err error) {
	self.channels = make(map[string](map[string]bool))
	return nil
}

func (self *Broadcast) HandleCast(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "JoinChannel" {
		self.handleJoinChannel(args[1].(string), args[2].(string))
	} else if method_name == "LeaveChannel" {
		self.handleLeaveChannel(args[1].(string), args[2].(string))
	} else if method_name == "Publish" {
		self.handlePublish(args[1].(*BroadcastMsg))
	}
}

func (self *Broadcast) HandleCall(args []interface{}) (interface{}, error) {
	return nil, nil
}

func (self *Broadcast) Terminate(reason string) (err error) {
	self.channels = nil
	return nil
}

/*
   Callback Handlers
*/

func (self *Broadcast) handleJoinChannel(playerId, channel string) {
	if v, ok := self.channels[channel]; ok {
		v[playerId] = true
	} else {
		m := map[string]bool{}
		m[playerId] = true
		self.channels[channel] = m
	}
}

func (self *Broadcast) handleLeaveChannel(playerId, channel string) {
	if v, ok := self.channels[channel]; ok {
		delete(v, playerId)
	}
}

func (self *Broadcast) handlePublish(msg *BroadcastMsg) {
	channel := msg.Channel
	if v, ok := self.channels[channel]; ok {
		for id, _ := range v {
			if _, ok := gen_server.GetGenServer(id); ok {
				gen_server.Cast(id, "broadcast", msg)
			} else {
				delete(self.channels[channel], id)
			}
		}
	}
}
