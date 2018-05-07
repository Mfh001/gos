package session_utils

import (
	"goslib/logger"
	"goslib/redisdb"
)

type Session struct {
	Uuid         string
	AccountId    string
	ServerId     string
	SceneId      string
	ConnectAppId string
	GameAppId    string
	Token        string
}

func Find(accountId string) (*Session, error) {
	uuid := "session:" + accountId
	sessionMap, err := redisdb.ServiceInstance().HGetAll(uuid).Result()
	if err != nil {
		logger.ERR("Find session failed: ", err)
		return nil, err
	}
	if len(sessionMap) == 0 {
		return nil, err
	}

	return parseSession(sessionMap), nil
}

func Create(params map[string]string) (*Session, error) {
	uuid := "session:" + params["accountId"]
	params["uuid"] = uuid
	setParams := make(map[string]interface{})
	for k, v := range params {
		setParams[k] = v
	}
	_, err := redisdb.ServiceInstance().HMSet(uuid, setParams).Result()
	if err != nil {
		logger.ERR("Create session failed: ", err)
	}
	return parseSession(params), err
}

func (self *Session) Save() error {
	params := make(map[string]interface{})
	params["accountId"] = self.AccountId
	params["serverId"] = self.ServerId
	params["sceneId"] = self.SceneId
	params["connectAppId"] = self.ConnectAppId
	params["gameAppId"] = self.GameAppId
	params["token"] = self.Token
	_, err := redisdb.ServiceInstance().HMSet(self.Uuid, params).Result()
	if err != nil {
		logger.ERR("Save session failed: ", err)
	}
	return err
}

func parseSession(params map[string]string) *Session {
	uuid := "session:" + params["accountId"]
	return &Session{
		Uuid:         uuid,
		AccountId:    params["accountId"],
		ServerId:     params["serverId"],
		SceneId:      params["sceneId"],
		ConnectAppId: params["connectAppId"],
		GameAppId:    params["gameAppId"],
		Token:        params["token"],
	}
}

func (self *Session) Del() error {
	_, err := redisdb.ServiceInstance().Del(self.Uuid).Result()
	if err != nil {
		logger.ERR("Del session failed: ", err)
	}
	return err
}
