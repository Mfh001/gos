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
package cache_mgr

import (
	"context"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Task struct {
	Content    string
	Version    int64
	NeedExpire bool
}

type Persister struct {
	queue         map[string]*Task
	persistTicker *time.Ticker
}

const SERVER = "__PERSISTER__"

func StartPersister() {
	gen_server.Start(SERVER, new(Persister))
}

var remainTask = &RemainTasksParams{}

func EnsurePersistered() {
	for {
		count, err := gen_server.Call(SERVER, remainTask)
		if err == nil && count.(int) == 0 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func persistToMySQL(playerId, content string, version int64, needExpire bool) {
	gen_server.Cast(SERVER, &PersistParams{playerId, &Task{
		Content:    content,
		Version:    version,
		NeedExpire: needExpire,
	}})
}

var ticker = &TickerPersistParams{}

func (self *Persister) Init(args []interface{}) (err error) {
	self.queue = make(map[string]*Task)
	self.persistTicker = time.NewTicker(time.Second)
	go func() {
		var err error
		for range self.persistTicker.C {
			_, err = gen_server.Call(SERVER, ticker)
			if err != nil {
				logger.ERR("persister tickerPersist failed: ", err)
			}
		}
	}()
	return nil
}

type PersistParams struct {
	playerId string
	task     *Task
}

func (self *Persister) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *PersistParams:
		self.queue[params.playerId] = params.task
		break
	}
}

type RemainTasksParams struct{}

func (self *Persister) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch req.Msg.(type) {
	case *TickerPersistParams:
		self.tickerPersist()
		break
	case *RemainTasksParams:
		return len(self.queue), nil
	}
	return nil, nil
}

func (self *Persister) Terminate(reason string) (err error) {
	return nil
}

type TickerPersistParams struct{}

func (self *Persister) tickerPersist() {
	for playerId, task := range self.queue {
		if err := persist(playerId, task); err == nil {
			delete(self.queue, playerId)
			if task.NeedExpire {
				_, err = redisdb.Instance().Expire(cacheKey(playerId), CACHE_EXPIRE).Result()
				if err != nil {
					logger.ERR("Persister setexpire failed: ", playerId, err)
				}
			}
		}
	}
}

func persist(playerId string, task *Task) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	upsert := true
	_, err = C.PlayerData.UpdateOne(ctx,
		bson.D{
			{"_id", playerId},
			{"UpdatedAt", bson.D{{"$lt", task.Version}}}},
		bson.D{
			{"$set", bson.D{
				{"_id", playerId},
				{"Content", task.Content},
				{"UpdatedAt", task.Version},
			}},
		}, &options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		logger.ERR("Persist data failed: ", playerId, err)
	}
	return
}
