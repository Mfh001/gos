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
package scene_mgr

import (
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/scene_utils"
	"sync"
)

var SceneLoadHandler func(scene *scene_utils.Scene) error = nil

/*
   GenServer Callbacks
*/
type SceneMgr struct {
}

const SERVER = "__scene_mgr__"

var loadedScenes *sync.Map

func Start() error {
	_, err := gen_server.Start(SERVER, new(SceneMgr))
	return err
}

func TryLoadScene(sceneId string) bool {
	if sceneId != "" {
		if loaded, ok := loadedScenes.Load(sceneId); !ok || !(loaded.(bool)) {
			success, err := gen_server.Call(SERVER, &LoadSceneParams{sceneId})
			if err != nil {
				logger.ERR("TryLoadScene failed: ", err)
				return false
			}
			return success.(bool)
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
		if SceneLoadHandler != nil {
			if err := doLoadScene(params.sceneId); err != nil {
				return false, err
			}
			loadedScenes.Store(params.sceneId, true)
			return true, nil
		} else {
			logger.ERR("SceneLoadHandler not exists: ", params.sceneId)
			return false, nil
		}
	}
	return false, nil
}

func (self *SceneMgr) Terminate(reason string) (err error) {
	return nil
}

func doLoadScene(sceneId string) error {
	scene, err := scene_utils.Find(sceneId)
	if err != nil {
		logger.ERR("LoadScene failed: ", sceneId, " err: ", err)
		return err
	}
	return SceneLoadHandler(scene)
}
