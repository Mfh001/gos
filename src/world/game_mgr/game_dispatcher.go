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
package game_mgr

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/scene_utils"
	"github.com/mafei198/gos/goslib/session_utils"
	"math"
	"math/rand"
	"sort"
	"time"
)

type DispatchInfo struct {
	AppId   string
	AppHost string
	AppPort string
}

func DispatchGame(accountId, sceneId string) (*DispatchInfo, error) {
	switch gosconf.SCENE_CONNECT_MODE {
	case gosconf.SCENE_CONNECT_MODE_PROXY:
		return DispatchByAccount(accountId, sceneId)
	case gosconf.SCENE_CONNECT_MODE_DIRECT:
		return DispatchByScene(accountId, sceneId)
	default:
		return nil, errors.New("please configure SCENE_CONNECT_MODE")
	}
}

func DispatchByScene(accountId, sceneId string) (*DispatchInfo, error) {
	game, err := DispatchScene(sceneId)
	if err != nil {
		return nil, err
	}
	if err := setGameAppIdToSession(accountId, game.Uuid, sceneId); err != nil {
		return nil, err
	}
	return &DispatchInfo{
		AppId:   game.Uuid,
		AppHost: game.Host,
		AppPort: game.Port,
	}, nil
}

func DispatchByAccount(accountId, sceneId string) (*DispatchInfo, error) {
	// make sure scene dispatched
	if _, err := DispatchScene(sceneId); err != nil {
		return nil, err
	}
	game, err := chooseGameApp()
	if err != nil {
		return nil, err
	}
	lockKey := "DispatchByAccount:" + accountId
	locked, err := redisdb.Instance().SetNX(lockKey, "uuid", 1*time.Second).Result()
	if locked {
		session, err := session_utils.Find(accountId)
		if session != nil && session.GameAppId != "" {
			return dispatchInfo(session.GameAppId)
		}
		err = setGameAppIdToSession(accountId, game.Uuid, sceneId)
		redisdb.Instance().Del(lockKey)
		if err != nil {
			return nil, err
		}
		return &DispatchInfo{
			AppId:   game.Uuid,
			AppHost: game.Host,
			AppPort: game.Port,
		}, nil
	} else {
		time.Sleep(10 * time.Millisecond)
		return DispatchByAccount(accountId, sceneId)
	}
}

func dispatchInfo(gameId string) (*DispatchInfo, error) {
	game, err := game_utils.Find(gameId)
	if err != nil {
		return nil, err
	}
	return &DispatchInfo{
		AppId:   game.Uuid,
		AppHost: game.Host,
		AppPort: game.Port,
	}, nil
}

func DispatchScene(sceneId string) (*game_utils.Game, error) {
	scene, err := scene_utils.Find(sceneId)
	if err != nil {
		return nil, err
	}

	if scene == nil {
		return nil, errors.New("scene not found")
	}

	if scene.GameAppId != "" {
		return game_utils.Find(scene.GameAppId)
	}

	game, err := chooseGameApp()
	if err != nil {
		return nil, err
	}

	lockKey := "lock_scene:" + sceneId
	locked, err := redisdb.Instance().SetNX(lockKey, "uuid", 1*time.Second).Result()
	if locked {
		if err := scene_utils.Update(scene.Uuid, "GameAppId", game.Uuid); err != nil {
			return nil, err
		} else {
			return game, nil
		}
	} else {
		time.Sleep(10 * time.Millisecond)
		return DispatchScene(sceneId)
	}
}

type GameCompare struct {
	Uuid  string
	Score int
}

func chooseGameApp() (*game_utils.Game, error) {
	games := make(map[string]*game_utils.Game)
	err := game_utils.LoadGames(games)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 {
		return nil, errors.New("no working Game")
	}

	list := make([]*GameCompare, len(games))
	idx := 0
	for _, game := range games {
		list[idx] = &GameCompare{
			Uuid:  game.Uuid,
			Score: gameAppScore(game),
		}
		idx++
	}

	// Sort games by score from best to bad
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	candidateList := make([]*GameCompare, 0)
	bestScore := list[0].Score

	for _, game := range list {
		if game.Score >= bestScore {
			candidateList = append(candidateList, game)
		}
	}

	// Choose random candidate
	randIdx := rand.Intn(len(candidateList))
	candidate := candidateList[randIdx]
	return games[candidate.Uuid], nil
}

func gameAppScore(game *game_utils.Game) int {
	// Ccu score: more remain space, higher score it get
	ccuScore := math.Max(float64(gosconf.GAME_CCU_MAX-game.Ccu)/float64(gosconf.GAME_CCU_MAX), 0)
	return int(ccuScore * 100)
}

func setGameAppIdToSession(accountId, gameAppId, sceneId string) error {
	session, err := session_utils.Find(accountId)
	if err != redis.Nil && err != nil {
		return err
	}
	if session == nil {
		session, err = session_utils.Create(&session_utils.Session{
			AccountId: accountId,
			GameAppId: gameAppId,
			SceneId:   sceneId,
		})
		return err
	} else {
		session.GameAppId = gameAppId
		session.SceneId = sceneId
		return session.Save()
	}
}
