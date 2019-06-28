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

package broadcast

import (
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
)

type BroadcastMsg struct {
	Category string
	Channel  string
	SenderId string
	Data     interface{}
}

type MsgHandler func(subscriber string, msg *BroadcastMsg)

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
	handler  MsgHandler
}

func (self *Broadcast) handleJoin(params *JoinParams) {
	self.subscribers[params.playerId] = params.handler
}

type LeaveParams struct{ playerId string }

func (self *Broadcast) handleLeave(params *LeaveParams) {
	if _, ok := self.subscribers[params.playerId]; ok {
		delete(self.subscribers, params.playerId)
	}
}

type PublishParams struct{ msg *BroadcastMsg }

func (self *Broadcast) handlePublish(params *PublishParams) {
	for subscriber, handler := range self.subscribers {
		handler(subscriber, params.msg)
	}
}
