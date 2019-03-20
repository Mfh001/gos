package game_mgr

import (
	"errors"
	"gosconf"
	"goslib/game_utils"
	"goslib/redisdb"
	"goslib/scene_utils"
	"goslib/session_utils"
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

/*
 * players with same sceneId will be dispatched to same game server by sceneId,
 * or will be dispatched to best score game server.
 */
func dispatchGame(accountId, sceneId string) (*DispatchInfo, error) {
	if sceneId != "" {
		return fastDispatch(accountId, sceneId)
	} else {
		return defaultDispatch(accountId)
	}
}

func fastDispatch(accountId, sceneId string) (*DispatchInfo, error) {
	scene, err := scene_utils.FindScene(sceneId)
	if err != nil {
		return nil, err
	}

	if scene == nil {
		scene, err = dispatchScene(sceneId)
		if err != nil {
			return nil, err
		}
		if scene == nil {
			return nil, errors.New("scene not found")
		}
	}

	err = setGameAppIdToSession(accountId, scene.GameAppId)
	if err != nil {
		return nil, err
	}

	game, err := game_utils.Find(scene.GameAppId)
	if err != nil {
		return nil, err
	}

	return &DispatchInfo{
		AppId:   game.Uuid,
		AppHost: game.Host,
		AppPort: game.Port,
	}, nil
}

func defaultDispatch(accountId string) (*DispatchInfo, error) {
	game, err := chooseGameApp()
	if err != nil {
		return nil, err
	}
	lockKey := "defaultDispatch:" + accountId
	locked, err := redisdb.Instance().SetNX(lockKey, "uuid", 10*time.Second).Result()
	if locked {
		session, err := session_utils.Find(accountId)
		if session != nil && session.GameAppId != "" {
			return dispatchInfo(session.GameAppId)
		}
		err = setGameAppIdToSession(accountId, game.Uuid)
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
		time.Sleep(1 * time.Second)
		session, err := session_utils.Find(accountId)
		if session != nil && session.GameAppId != "" {
			return dispatchInfo(session.GameAppId)
		}
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

func dispatchScene(sceneId string) (*scene_utils.Scene, error) {
	game, err := chooseGameApp()
	if err != nil {
		return nil, err
	}
	uuid := scene_utils.GenUuid(sceneId)
	lockKey := "lock_scene:" + uuid
	locked, err := redisdb.Instance().SetNX(lockKey, "uuid", 10*time.Second).Result()
	if locked {
		scene, err := scene_utils.CreateScene(&scene_utils.Scene{
			Uuid:      scene_utils.GenUuid(sceneId),
			GameAppId: game.Uuid,
		})
		redisdb.Instance().Del(lockKey)
		return scene, err
	} else {
		time.Sleep(1 * time.Second)
		scene, err := scene_utils.FindScene(sceneId)
		return scene, err
	}
}

type GameCompare struct {
	Uuid  string
	Score int
}

func chooseGameApp() (*game_utils.Game, error) {
	var games map[string]*game_utils.Game
	err := game_utils.LoadGames(games)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 {
		return nil, errors.New("no working Game")
	}

	list := make([]*GameCompare, len(games))
	for _, game := range games {
		list[len(list)] = &GameCompare{
			Uuid:  game.Uuid,
			Score: gameAppScore(game),
		}
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

func setGameAppIdToSession(accountId string, gameAppId string) error {
	session, err := session_utils.Find(accountId)
	if err != nil {
		return err
	}
	if session == nil {
		session, err = session_utils.Create(&session_utils.Session{
			AccountId: accountId,
			GameAppId: gameAppId,
		})
		return err
	} else {
		session.GameAppId = gameAppId
		return session.Save()
	}
}
