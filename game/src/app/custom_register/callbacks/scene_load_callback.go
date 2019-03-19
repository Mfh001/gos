package callbacks

import (
	"goslib/logger"
	"goslib/scene_mgr"
)

const (
	SCENE_SERVER_DEFAULT = "SCENE_SERVER_DEFAULT"
)

func RegisterSceneLoad() {
	scene_mgr.SceneLoadHandler = func(sceneId string, sceneType string, sceneConfigId string) {
		logger.INFO("LoadSceneCallback sceneType: ", sceneType, " sceneConfigId: ", sceneConfigId)
		switch sceneType {
		case SCENE_SERVER_DEFAULT:
			logger.INFO("Need load default server scene")
		}
	}
}
