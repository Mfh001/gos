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
	"container/list"
	"errors"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
	"time"
)

type Persister struct {
	server        *gen_server.GenServer
	tasks         *list.List
	persistTicker *time.Ticker
}

var ticker = &TickerPersistParams{}

func (self *Persister) Init(args []interface{}) (err error) {
	self.tasks = list.New()
	self.persistTicker = time.NewTicker(time.Second)
	go func() {
		var err error
		for range self.persistTicker.C {
			if self.server != nil {
				_, err = self.server.Call(ticker)
				if err != nil {
					logger.ERR("shared_data_persist tickerPersist failed: ", err)
				}
			}
		}
	}()
	return nil
}

func (self *Persister) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch req.Msg.(type) {
	case *TickerPersistParams:
		self.tickerPersist()
		break
	}
	return nil, nil
}

func (self *Persister) HandleCast(req *gen_server.Request) {
	self.tasks.PushBack(req.Msg)
	flushQueue(self.tasks)
}

func (self *Persister) Terminate(reason string) (err error) {
	if size := flushQueue(self.tasks); size == 0 {
		self.persistTicker.Stop()
		return nil
	} else {
		return errors.New("share_data_persist failed")
	}
}

type TickerPersistParams struct{}

func (self *Persister) tickerPersist() {
	flushQueue(self.tasks)
}

func flushQueue(tasks *list.List) int {
	for task := tasks.Front(); task != nil; task = tasks.Front() {
		if err := executeCmd(task.Value); err != nil {
			logger.ERR("share_data persist failed: ", err)
			break
		} else {
			tasks.Remove(task)
		}
	}
	return tasks.Len()
}

type PersistParams struct {
	handler WriteHandler
	table   string
	key     string
	rec     interface{}
}

type RemoveParams struct {
	handler DeleteHandler
	table   string
	key     string
}

func executeCmd(msg interface{}) error {
	switch params := msg.(type) {
	case *PersistParams:
		return params.handler(params.table, params.key, params.rec)
	case *RemoveParams:
		return params.handler(params.table, params.key)
	}
	return nil
}
