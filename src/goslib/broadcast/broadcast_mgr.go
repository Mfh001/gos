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
