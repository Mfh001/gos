package game_utils

import (
	"goslib/redisdb"
	"goslib/logger"
	"strconv"
	"gosconf"
)

type Game struct {
	Uuid string
	Host string
	Port string
	Ccu  int32
	CcuMax int32
	ActiveAt int64
}

func Find(uuid string) (*Game, error) {
	valueMap, err := redisdb.Instance().HGetAll(uuid).Result()
	if err != nil {
		logger.ERR("Find game failed: ", err)
		return nil, err
	}
	if len(valueMap) == 0 {
		return nil, err
	}

	return parseObject(valueMap), nil
}

func LoadAll(apps map[string]*Game) error {
	ids, err := redisdb.Instance().SMembers(gosconf.RK_GAME_APP_IDS).Result()
	if err != nil {
		logger.ERR("Redis load games failed: ", err)
		return err
	}
	for _, id := range ids {
		if app, _ := Find(id); app != nil{
			apps[id] = app
		}
	}
	return nil
}

func Create(params map[string]string) (*Game, error) {
	setParams := make(map[string]interface{})
	for k,v := range params  {
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
	params["host"] = self.Host
	params["port"] = self.Port
	params["ccu"] = self.Ccu
	params["ccuMax"] = self.CcuMax
	params["activeAt"] = self.ActiveAt
	_, err := redisdb.Instance().HMSet(self.Uuid, params).Result()
	if err != nil {
		logger.ERR("Save game failed: ", err)
	}
	return err
}

func parseObject(params map[string]string) *Game {
	Ccu, _ := strconv.Atoi(params["ccu"])
	CcuMax, _ := strconv.Atoi(params["ccuMax"])
	ActiveAt, _ := strconv.Atoi(params["activeAt"])
	return &Game{
		Uuid: params["uuid"],
		Host: params["host"],
		Port: params["port"],
		Ccu: int32(Ccu),
		CcuMax: int32(CcuMax),
		ActiveAt: int64(ActiveAt),
	}
}

func (self *Game) Del() error {
	_, err := redisdb.Instance().Del(self.Uuid).Result()
	if err != nil {
		logger.ERR("Del game failed: ", err)
	}
	return err
}

