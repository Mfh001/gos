package scene_utils

import (
	"goslib/redisdb"
	"goslib/logger"
	"strconv"
	"gosconf"
)

type Scene struct {
	Uuid string
	GameAppId string
	SceneType string
	SceneConfigId string
	Ccu int
	CcuMax int
}

type SceneConf struct {
	ConfId string
	SceneType string
	CcuMax int
}

func LoadScenes() ([]*Scene, map[string]*Scene) {
	mapScenes := make(map[string]*Scene)
	ids, _ := redisdb.Instance().SMembers(gosconf.RK_SCENE_IDS).Result()
	scenes := make([]*Scene, 0)
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

func FindScene(sceneId string) (*Scene, error) {
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


func CreateDefaultServerScene(serverId string, conf *SceneConf) (*Scene, error) {
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
	return &Scene{
		Uuid: serverId,
		GameAppId: "",
		SceneType: conf.SceneType,
		SceneConfigId: conf.ConfId,
		Ccu: 0,
		CcuMax: conf.CcuMax,
	}, nil
}

func parseScene(valueMap map[string]string) *Scene {
	if valueMap["uuid"] == "" {
		return nil
	}
	ccu, _ := strconv.Atoi(valueMap["ccu"])
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	return &Scene{
		Uuid: valueMap["uuid"],
		GameAppId: valueMap["gameAppId"],
		SceneType: valueMap["sceneType"],
		SceneConfigId: valueMap["sceneConfigId"],
		Ccu: ccu,
		CcuMax: ccuMax,
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

