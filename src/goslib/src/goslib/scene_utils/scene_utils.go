package scene_utils

import (
	"github.com/go-redis/redis"
	"gosconf"
	"goslib/logger"
	"goslib/redisdb"
)

type Scene struct {
	Uuid      string
	GameAppId string
}

func GenUuid(sceneId string) string {
	uuid := "server_scene:" + sceneId
	return uuid
}

func LoadScenes(mapScenes map[string]*Scene) {
	ids, _ := redisdb.Instance().SMembers(gosconf.RK_SCENE_IDS).Result()
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		if id == "" {
			continue
		}
		valueMap, _ := redisdb.Instance().HGetAll(id).Result()
		app := parseScene(valueMap)
		mapScenes[app.Uuid] = app
	}
	logger.INFO("idSize: ", len(ids), " appSize: ", len(mapScenes))
}

func FindScene(sceneId string) (*Scene, error) {
	sceneMap, err := redisdb.Instance().HGetAll(sceneId).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		logger.ERR("findScene: ", sceneId, " failed: ", err)
		return nil, err
	}
	return parseScene(sceneMap), nil
}

func CreateScene(scene *Scene) (*Scene, error) {
	err := scene.Save()
	if err != nil {
		logger.ERR("Create session failed: ", err)
	} else {
		redisdb.Instance().SAdd(gosconf.RK_SCENE_IDS, scene.Uuid)
	}
	return scene, err
}

func (self *Scene) Save() error {
	params := make(map[string]interface{})
	params["uuid"] = self.Uuid
	params["gameAppId"] = self.GameAppId
	_, err := redisdb.Instance().HMSet(self.Uuid, params).Result()
	if err != nil {
		logger.ERR("Save game failed: ", err)
	}
	return err
}

func parseScene(valueMap map[string]string) *Scene {
	if valueMap["uuid"] == "" {
		return nil
	}
	return &Scene{
		Uuid:      valueMap["uuid"],
		GameAppId: valueMap["gameAppId"],
	}
}
