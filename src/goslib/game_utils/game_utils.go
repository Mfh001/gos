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
package game_utils

import (
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"strconv"
)

type Game struct {
	Uuid       string
	Role       string
	Host       string
	Port       string
	RpcHost    string
	RpcPort    string
	StreamPort string
	Ccu        int32
	CcuMax     int32
	ActiveAt   int64
}

func Find(uuid string) (*Game, error) {
	valueMap, err := redisdb.Instance().HGetAll(uuid).Result()
	if err == nil && len(valueMap) == 0 {
		return nil, redis.Nil
	}
	if err != nil {
		logger.ERR("Find game failed: ", err)
		return nil, err
	}

	return parseObject(valueMap), nil
}

func LoadGames(apps map[string]*Game) error {
	ids, err := redisdb.Instance().SMembers(gosconf.RK_GAME_APP_IDS).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		logger.ERR("Redis load games failed: ", err)
		return err
	}
	for _, id := range ids {
		if app, _ := Find(id); app != nil {
			apps[id] = app
		}
	}
	return nil
}

func Create(params map[string]string) (*Game, error) {
	setParams := make(map[string]interface{})
	for k, v := range params {
		setParams[k] = v
	}
	_, err := redisdb.Instance().HMSet(params["uuid"], setParams).Result()
	if err != nil {
		logger.ERR("Create game failed: ", err)
	}
	return parseObject(params), err
}

func (self *Game) Save() error {
	params := make(map[string]interface{})
	params["Uuid"] = self.Uuid
	params["Role"] = self.Role
	params["Host"] = self.Host
	params["Port"] = self.Port
	params["RpcHost"] = self.RpcHost
	params["RpcPort"] = self.RpcPort
	params["StreamPort"] = self.StreamPort
	params["Ccu"] = self.Ccu
	params["CcuMax"] = self.CcuMax
	params["ActiveAt"] = self.ActiveAt
	_, err := redisdb.Instance().HMSet(self.Uuid, params).Result()
	if err != nil {
		logger.ERR("Save game failed: ", err)
	}
	return err
}

func parseObject(params map[string]string) *Game {
	Ccu, _ := strconv.Atoi(params["Ccu"])
	CcuMax, _ := strconv.Atoi(params["CcuMax"])
	ActiveAt, _ := strconv.Atoi(params["ActiveAt"])
	return &Game{
		Uuid:       params["Uuid"],
		Role:       params["Role"],
		Host:       params["Host"],
		Port:       params["Port"],
		RpcHost:    params["RpcHost"],
		RpcPort:    params["RpcPort"],
		StreamPort: params["StreamPort"],
		Ccu:        int32(Ccu),
		CcuMax:     int32(CcuMax),
		ActiveAt:   int64(ActiveAt),
	}
}

func (self *Game) Del() error {
	_, err := redisdb.Instance().Del(self.Uuid).Result()
	if err != nil {
		logger.ERR("Del game failed: ", err)
	}
	return err
}
