package game_mgr

import (
	"goslib/redisdb"
	"goslib/logger"
	"strconv"
	"gosconf"
)

func LoadScenes() ([]*SceneCell, map[string]*SceneCell) {
	mapScenes := make(map[string]*SceneCell)
	ids, _ := redisdb.Instance().SMembers(gosconf.RK_SCENE_IDS).Result()
	scenes := make([]*SceneCell, 0)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		if id == "" {
			continue
		}
		valueMap, _ := redisdb.Instance().HGetAll(id).Result()
		app := parseScene(valueMap)
		scenes = append(scenes, app)
		mapScenes[app.Uuid] = app
	}
	logger.INFO("idSize: ", len(ids), " appSize: ", len(scenes))
	return scenes, mapScenes
}

func FindScene(sceneId string) (*SceneCell, error) {
	sceneMap, err := redisdb.Instance().HGetAll(sceneId).Result()
	if err != nil {
		logger.ERR("findScene: ", sceneId, " failed: ", err)
		return nil, err
	}
	if len(sceneMap) == 0 {
		return nil, nil
	}
	return parseScene(sceneMap), nil
}

func FindSceneConf(confId string) (*SceneConf, error) {
	valueMap, err := redisdb.Instance().HGetAll(confId).Result()
	if err != nil {
		logger.ERR("findSceneConf: ", confId, " failed: ", err)
		return nil, err
	}
	if len(valueMap) == 0 {
		logger.ERR("SceneConf: ", confId, " Not Found!")
		//FIXME
		return &SceneConf{
			ConfId: "default_conf_id",
			SceneType: "default_server",
			CcuMax: 100,
		}, nil
		//return nil, nil
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
	_, err := redisdb.Instance().HMSet(serverId, params).Result()
	if err != nil {
		logger.ERR("createDefaultServerScene: ", serverId, " failed: ", err)
		return nil, err
	}
	redisdb.Instance().SAdd(gosconf.RK_SCENE_IDS, serverId)
	return &SceneCell{
		Uuid: serverId,
		GameAppId: "",
		SceneType: conf.SceneType,
		SceneConfigId: conf.ConfId,
		Ccu: 0,
		CcuMax: conf.CcuMax,
		ServedServers: make([]string, 0),
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
		Uuid: valueMap["uuid"],
		GameAppId: valueMap["gameAppId"],
		SceneType: valueMap["sceneType"],
		SceneConfigId: valueMap["sceneConfigId"],
		Ccu: ccu,
		CcuMax: ccuMax,
		ServedServers: servedServers,
	}
}

func parseSceneConf(valueMap map[string]string) *SceneConf {
	if valueMap["uuid"] == "" {
		return nil
	}
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	return &SceneConf{
		ConfId: valueMap["confId"],
		SceneType: valueMap["SceneType"],
		CcuMax: ccuMax,
	}
}

