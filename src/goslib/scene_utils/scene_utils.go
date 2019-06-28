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
package scene_utils

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/utils/redis_utils"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"time"
)

type Scene struct {
	Uuid       string
	Name       string
	Category   int32
	ConfId     int32
	Registered int32
	GameAppId  string
	CreatedAt  int64
}

func GenUuid(name string) string {
	return "{scene}." + name
}

// Add scene from admin tools
func Add(name string, confId, category int32) (*Scene, error) {
	uuid := GenUuid(name)
	exist, err := redisdb.Instance().SIsMember(gosconf.RK_SCENE_IDS, uuid).Result()
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("scene already exist")
	}
	scene := &Scene{
		Uuid:       uuid,
		Name:       name,
		Category:   category,
		ConfId:     confId,
		Registered: 0,
		GameAppId:  "",
		CreatedAt:  time.Now().Unix(),
	}
	err = scene.save()
	if err != nil {
		logger.ERR("Create session failed: ", err)
	} else {
		redisdb.Instance().SAdd(gosconf.RK_SCENE_IDS, scene.Uuid)
	}
	return scene, err
}

func LoadAll() ([]*Scene, error) {
	ids, err := redisdb.Instance().SMembers(gosconf.RK_SCENE_IDS).Result()
	if err != nil {
		return nil, err
	}
	scenes := make([]*Scene, 0)
	for _, id := range ids {
		scene, err := Find(id)
		if err != nil {
			return nil, err
		}
		scenes = append(scenes, scene)
	}
	return scenes, nil
}

func Find(sceneId string) (*Scene, error) {
	sceneMap, err := redisdb.Instance().HGetAll(sceneId).Result()
	if err == redis.Nil || len(sceneMap) == 0 {
		return nil, nil
	}
	if err != nil {
		logger.ERR("findScene: ", sceneId, " failed: ", err)
		return nil, err
	}
	return parseScene(sceneMap), nil
}

func Update(uuid, k, v string) error {
	_, err := redisdb.Instance().HSet(uuid, k, v).Result()
	if err != nil {
		logger.ERR("Save scene failed: ", err)
	}
	return err
}

func (self *Scene) save() error {
	pipe := redisdb.Instance().TxPipeline()
	params := make(map[string]interface{})
	params["Uuid"] = self.Uuid
	params["Name"] = self.Name
	params["Category"] = self.Category
	params["ConfId"] = self.ConfId
	params["Registered"] = self.Registered
	params["GameAppId"] = self.GameAppId
	params["CreatedAt"] = self.CreatedAt
	pipe.HMSet(self.Uuid, params)
	pipe.SAdd(gosconf.RK_SCENE_IDS, self.Uuid)
	if _, err := pipe.Exec(); err != nil {
		logger.ERR("Save scene failed: ", err)
	}
	return nil
}

func parseScene(valueMap map[string]string) *Scene {
	if valueMap["Uuid"] == "" {
		return nil
	}
	return &Scene{
		Uuid:       valueMap["Uuid"],
		Name:       valueMap["Name"],
		Category:   redis_utils.ToInt32(valueMap["Category"]),
		ConfId:     redis_utils.ToInt32(valueMap["ConfId"]),
		Registered: redis_utils.ToInt32(valueMap["Registered"]),
		GameAppId:  valueMap["GameAppId"],
		CreatedAt:  redis_utils.ToInt64(valueMap["CreatedAt"]),
	}
}
