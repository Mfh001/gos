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
package session_utils

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/account_utils"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/secure"
	"google.golang.org/grpc"
	"net"
)

type Session struct {
	Uuid      string
	AccountId string
	Token     string // Token which identify player authed
	GameAppId string // GameServer which served scenes
	SceneId   string // Scene is logic space, the players with same sceneId stay in same GameServer
}

var GameMgrRpcClient proto.GameDispatcherClient

func Start() {
	connectGameMgr()
}

func GetPlayerSession(accountId string) (*Session, error) {
	session, err := Find(accountId)
	if err == redis.Nil {
		account, err := account_utils.Lookup(accountId)
		if err != nil {
			return nil, err
		}
		return ChooseGameServer(accountId, account.GroupId)
	}
	return session, err
}

func Find(accountId string) (*Session, error) {
	uuid := key(accountId)
	sessionMap, err := redisdb.Instance().HGetAll(uuid).Result()
	if err == nil && len(sessionMap) == 0 {
		err = redis.Nil
	}
	if err != nil {
		//logger.ERR("Find session failed: ", accountId, err)
		return nil, err
	}

	return parseSession(sessionMap), nil
}

func key(accountId string) string {
	return "session:" + accountId
}

func Active(accountId string) {
	redisdb.Instance().Expire(key(accountId), gosconf.SESSION_EXPIRE_DURATION)
}

func Create(session *Session) (*Session, error) {
	uuid := key(session.AccountId)
	session.Uuid = uuid
	return session, session.Save()
}

func (self *Session) Save() error {
	params := make(map[string]interface{})
	params["Uuid"] = self.Uuid
	params["AccountId"] = self.AccountId
	params["SceneId"] = self.SceneId
	params["GameAppId"] = self.GameAppId
	params["Token"] = self.Token
	logger.INFO("save session: ")
	for k, v := range params {
		logger.INFO(k, ":", v.(string))
	}
	_, err := redisdb.Instance().HMSet(self.Uuid, params).Result()
	if err != nil {
		logger.ERR("Save session failed: ", err)
	}
	return err
}

func parseSession(params map[string]string) *Session {
	uuid := key(params["AccountId"])
	return &Session{
		Uuid:      uuid,
		AccountId: params["AccountId"],
		SceneId:   params["SceneId"],
		GameAppId: params["GameAppId"],
		Token:     params["Token"],
	}
}

func (self *Session) Del() error {
	_, err := redisdb.Instance().Del(self.Uuid).Result()
	if err != nil {
		logger.ERR("Del session failed: ", err)
	}
	return err
}

func connectGameMgr() {
	conf := gosconf.RPC_FOR_GAME_APP_MGR
	conn, err := grpc.Dial(net.JoinHostPort(gosconf.GetWorldIP(), conf.ListenPort), conf.DialOptions...)
	if err != nil {
		logger.ERR("connectGameMgr failed: ", err)
		return
	}

	GameMgrRpcClient = proto.NewGameDispatcherClient(conn)
}

// Request GameAppMgr to dispatch GameApp for session
func ChooseGameServer(accountId, sceneId string) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gosconf.RPC_REQUEST_TIMEOUT)
	defer cancel()

	_, err := GameMgrRpcClient.DispatchGame(ctx, &proto.DispatchGameRequest{
		AccountId: accountId,
		SceneId:   sceneId,
	})
	if err != nil {
		logger.ERR("DispatchGame failed: ", err)
		return nil, err
	}

	session, err := Find(accountId)
	if err != nil {
		logger.ERR("Dispatch account find session failed: ", err)
		return nil, err
	}
	session.Token = secure.SessionToken()
	if err = session.Save(); err != nil {
		return nil, err
	}

	return session, nil
}
