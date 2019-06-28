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
package shared_data

import (
	"github.com/mafei198/gos/goslib/gen_server"
	"sync"
	"time"
)

type Manager struct {
}

var clients = &sync.Map{}
var server = "__shared_data_mgr__"

func StartMgr() {
	_, err := gen_server.Start(server, &Manager{})
	if err != nil {
		panic(err)
	}
}

func EnsureShutdown() {
	for {
		finished := true
		clients.Range(func(serverId, client interface{}) bool {
			if err := client.(*SharedData).Stop("shutdown"); err != nil {
				finished = false
				return false
			}
			return true
		})
		if finished {
			return
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func Client(name string) *SharedData {
	if client, ok := clients.Load(name); ok {
		return client.(*SharedData)
	}
	return nil
}

func Instance(options *StartOptions) (*SharedData, error) {
	if client, ok := clients.Load(options.Name); ok {
		return client.(*SharedData), nil
	}
	value, err := gen_server.Call(server, &StartClientParams{options})
	if err != nil {
		return nil, err
	}
	return value.(*SharedData), nil
}

func (self *Manager) Init(args []interface{}) (err error) {
	return nil
}

type StartClientParams struct {
	options *StartOptions
}

func (self *Manager) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *StartClientParams:
		if client, ok := clients.Load(params.options.Name); ok {
			return client.(*SharedData), nil
		}
		client := &SharedData{
			options: params.options,
		}
		if _, err := gen_server.Start(params.options.Name, client); err != nil {
			return nil, err
		}
		clients.Store(params.options.Name, client)
		return client, nil
	}
	return nil, nil
}

func (self *Manager) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *TaskParams:
		self.handleTask(params)
		break
	}
}

type TaskParams struct{}

func (self *Manager) handleTask(params *TaskParams) {
}

func (self *Manager) Terminate(reason string) (err error) {
	return nil
}
