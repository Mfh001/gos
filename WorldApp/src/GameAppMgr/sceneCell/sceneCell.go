package sceneCell

import (
	"goslib/redisDB"
	"goslib/logger"
	. "GameAppMgr/commonConst"
	"strconv"
	"gosconf"
)

func LoadScenes() ([]*SceneCell, map[string]*SceneCell) {
	mapScenes := make(map[string]*SceneCell)
	ids, _ := redisDB.Instance().SMembers(gosconf.RK_SCENE_IDS).Result()
	scenes := make([]*SceneCell, 0)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		valueMap, _ := redisDB.Instance().HGetAll(id).Result()
		app := parseScene(valueMap)
		scenes = append(scenes, app)
		mapScenes[app.Uuid] = app
	}
	logger.INFO("idSize: ", len(ids), " appSize: ", len(scenes))
	return scenes, mapScenes
}

func FindScene(sceneId string) (*SceneCell, error) {
	sceneMap, err := redisDB.Instance().HGetAll(sceneId).Result()
	if err != nil {
		logger.ERR("findScene: ", sceneId, " failed: ", err)
		return nil, err
	}
	return parseScene(sceneMap), nil
}

func FindSceneConf(confId string) (*SceneConf, error) {
	valueMap, err := redisDB.Instance().HGetAll(confId).Result()
	if err != nil {
		logger.ERR("findSceneConf: ", confId, " failed: ", err)
		return nil, err
	}
	return parseSceneConf(valueMap), nil
}


func CreateDefaultServerScene(serverId string, conf *SceneConf) (*SceneCell, error) {
	params := make(map[string]interface{})
	params["uuid"] = serverId
	params["gameAppId"] = ""
	params["sceneType"] = conf.SceneType
	params["sceneConfigId"] = conf.ConfId
	params["ccu"] = 0
	params["ccuMax"] = conf.CcuMax
	params["servedServers"] = ""
	_, err := redisDB.Instance().HMSet(serverId, params).Result()
	if err != nil {
		logger.ERR("createDefaultServerScene: ", serverId, " failed: ", err)
		return nil, err
	}
	return &SceneCell{
		serverId,
		"",
		conf.SceneType,
		conf.ConfId,
		0,
		conf.CcuMax,
		make([]string, 0),
	}, nil
}

func parseScene(valueMap map[string]string) *SceneCell {
	if valueMap["uuid"] == "" {
		return nil
	}
	ccu, _ := strconv.Atoi(valueMap["ccu"])
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	servedServers := make([]string, 0)
	return &SceneCell{
		valueMap["uuid"],
		valueMap["gameAppId"],
		valueMap["sceneType"],
		valueMap["sceneConfigId"],
		ccu,
		ccuMax,
		servedServers,
	}
}

func parseSceneConf(valueMap map[string]string) *SceneConf {
	if valueMap["uuid"] == "" {
		return nil
	}
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	return &SceneConf{
		valueMap["confId"],
		valueMap["SceneType"],
		ccuMax,
	}
}

