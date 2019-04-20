package session_utils

import (
	"github.com/go-redis/redis"
	"goslib/logger"
	"goslib/redisdb"
)

type Session struct {
	Uuid      string
	AccountId string
	Token     string // Token which identify player authed

	ServerId string

	GameRole  string
	GameAppId string // GameServer which served scenes
	SceneId   string // Scene is logic space, the players with same sceneId stay in same GameServer

	RoomAppId string // RoomServer which served rooms
	RoomId    string // Room which served player ROOM logic
}

func Find(accountId string) (*Session, error) {
	uuid := "session:" + accountId
	sessionMap, err := redisdb.Instance().HGetAll(uuid).Result()
	if err == nil && len(sessionMap) == 0 {
		err = redis.Nil
	}
	if err != nil {
		logger.ERR("Find session failed: ", accountId, err)
		return nil, err
	}

	return parseSession(sessionMap), nil
}

func Create(session *Session) (*Session, error) {
	uuid := "session:" + session.AccountId
	session.Uuid = uuid
	return session, session.Save()
}

func (self *Session) Save() error {
	params := make(map[string]interface{})
	params["Uuid"] = self.Uuid
	params["AccountId"] = self.AccountId
	params["ServerId"] = self.ServerId
	params["SceneId"] = self.SceneId
	params["GameRole"] = self.GameRole
	params["GameAppId"] = self.GameAppId
	params["RoomAppId"] = self.RoomAppId
	params["RoomId"] = self.RoomId
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
	uuid := "session:" + params["AccountId"]
	return &Session{
		Uuid:      uuid,
		AccountId: params["AccountId"],
		ServerId:  params["ServerId"],
		SceneId:   params["SceneId"],
		GameRole:  params["GameRole"],
		GameAppId: params["GameAppId"],
		RoomAppId: params["RoomAppId"],
		RoomId:    params["RoomId"],
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
