/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
 */
package callbacks

import (
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/scene_mgr"
	"github.com/mafei198/gos/goslib/scene_utils"
	"time"
)

const (
	SCENE_MAP int32 = 1
)

func RegisterSceneLoad() {
	scene_mgr.SceneLoadHandler = func(scene *scene_utils.Scene) error {
		logger.INFO("LoadSceneCallback category: ", scene.Category, " confId: ", scene.ConfId)
		switch scene.Category {
		case SCENE_MAP:
			return nil
		}
		return nil
	}
}

func markSceneLoaded(sceneId string) {
	_, err := redisdb.Instance().SAdd(gosconf.INITED_MAP_SCENE_IDS, sceneId).Result()
	if err != nil {
		logger.ERR("markSceneLoaded failed: ", sceneId)
		time.Sleep(1 * time.Second)
		markSceneLoaded(sceneId)
	}
}
