package scene_mgr

import (
	"github.com/go-redis/redis"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/redisdb"
	"sync"
)

var SceneLoadHandler func(sceneId string, sceneType string, sceneConfigId string) = nil

/*
   GenServer Callbacks
*/
type SceneMgr struct {
}

const SERVER = "__scene_mgr__"

var loadedScenes *sync.Map

func Start() {
	gen_server.Start(SERVER, new(SceneMgr))
}

func TryLoadScene(sceneId string) bool {
	if sceneId != "" {
		if loaded, ok := loadedScenes.Load(sceneId); !ok || !(loaded.(bool)) {
			gen_server.Call(SERVER, &LoadSceneParams{sceneId})
		}
	}
	return true
}

func (self *SceneMgr) Init(args []interface{}) (err error) {
	loadedScenes = &sync.Map{}
	return nil
}

func (self *SceneMgr) HandleCast(req *gen_server.Request) {
}

type LoadSceneParams struct {
	sceneId string
}

func (self *SceneMgr) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *LoadSceneParams:
		doLoadScene(params.sceneId)
		loadedScenes.Store(params.sceneId, true)
		return true, nil
	}
	return nil, nil
}

func (self *SceneMgr) Terminate(reason string) (err error) {
	return nil
}

func doLoadScene(sceneId string) {
	valueMap, err := redisdb.Instance().HGetAll(sceneId).Result()
	if err == redis.Nil || len(valueMap) == 0 {
		logger.ERR("LoadScene failed: scene not found ", sceneId)
		return
	}
	if err != nil {
		logger.ERR("LoadScene failed: ", sceneId, " err: ", err)
		return
	}
	sceneType := valueMap["sceneType"]
	sceneConfigId := valueMap["sceneConfigId"]
	SceneLoadHandler(sceneId, sceneType, sceneConfigId)
}
