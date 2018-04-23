package gameCell

import (
	"goslib/redisDB"
	"strconv"
	"goslib/logger"
	."GameAppMgr/commonConst"
)

func SetupForTest() {
	// GameCell
	for i := 0; i < 10; i++  {
		num := strconv.Itoa(i)
		uuid := "fake_game_app:" + num
		value := make(map[string]interface{})
		value["uuid"] = uuid
		value["name"] = "agent:" + num
		value["host"] = "127.0.0.1"
		value["port"] = "500" + num
		value["ccu"] = 0
		value["ccuMax"] = 100
		value["status"] = SERVER_STATUS_WORKING
		redisDB.Instance().HMSet(uuid, value)
		redisDB.Instance().SAdd(GAME_APP_IDS_KEY, uuid)
	}
	// SceneConf
	for i := 0; i < 10; i++  {
		num := strconv.Itoa(i)
		uuid := "fake_scene_conf:" + num
		value := make(map[string]interface{})
		value["confId"] = uuid
		value["sceneType"] = "agent:" + num
		value["ccuMax"] = 100
		redisDB.Instance().HMSet(uuid, value)
		redisDB.Instance().SAdd(SCENE_CONF_IDS_KEY, uuid)
	}
}

func LoadApps() ([]*GameCell, map[string]*GameCell) {
	mapApps := make(map[string]*GameCell)
	ids, _ := redisDB.Instance().SMembers(GAME_APP_IDS_KEY).Result()
	apps := make([]*GameCell, 0)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		valueMap, _ := redisDB.Instance().HGetAll(id).Result()
		app := parseGameApp(valueMap)
		apps = append(apps, app)
		mapApps[app.Uuid] = app
	}
	logger.INFO("idSize: ", len(ids), " appSize: ", len(apps))
	return apps, mapApps
}

func FindGameApp(gameAppId string) (*GameCell, error) {
	valueMap, err := redisDB.Instance().HGetAll(gameAppId).Result()
	if err != nil {
		logger.ERR("findGameApp: ", gameAppId, " failed: ", err)
		return nil, err
	}
	return parseGameApp(valueMap), nil
}

func parseGameApp(valueMap map[string]string) *GameCell {
	if valueMap["uuid"] == "" {
		return nil
	}
	ccu, _ := strconv.Atoi(valueMap["ccu"])
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	status, _ := strconv.Atoi(valueMap["status"])
	servedScenes := make([]string, 0)
	return &GameCell{
		valueMap["uuid"],
		valueMap["name"],
		valueMap["host"],
		valueMap["port"],
		ccu,
		ccuMax,
		status,
		servedScenes,
	}
}
