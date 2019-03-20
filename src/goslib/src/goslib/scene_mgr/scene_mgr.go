package scene_mgr

import (
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
			gen_server.Call(SERVER, "LoadScene", sceneId)
		}
	}
	return true
}

func (self *SceneMgr) Init(args []interface{}) (err error) {
	loadedScenes = &sync.Map{}
	return nil
}

func (self *SceneMgr) HandleCast(args []interface{}) {
}

func (self *SceneMgr) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "LoadScene" {
		sceneId := args[1].(string)
		doLoadScene(sceneId)
		loadedScenes.Store(sceneId, true)
		return true, nil
	}
	return nil, nil
}

func (self *SceneMgr) Terminate(reason string) (err error) {
	return nil
}

func doLoadScene(sceneId string) {
	valueMap, err := redisdb.Instance().HGetAll(sceneId).Result()
	if err != nil {
		logger.ERR("LoadScene failed: ", sceneId, " err: ", err)
		return
	}
	sceneType := valueMap["sceneType"]
	sceneConfigId := valueMap["sceneConfigId"]
	SceneLoadHandler(sceneId, sceneType, sceneConfigId)
}
