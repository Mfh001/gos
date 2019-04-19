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
func DispatchGame(role, accountId, sceneId string) (*DispatchInfo, error) {
	if sceneId != "" {
		return fastDispatch(role, accountId, sceneId)
	} else {
		return defaultDispatch(role, accountId, sceneId)
	}
}

func fastDispatch(role, accountId, sceneId string) (*DispatchInfo, error) {
	scene, err := scene_utils.FindScene(sceneId)
	if err != nil {
		return nil, err
	}

	if scene == nil {
		scene, err = dispatchScene(role, sceneId)
		if err != nil {
			return nil, err
		}
		if scene == nil {
			return nil, errors.New("scene not found")
		}
	}

	err = setGameAppIdToSession(role, accountId, scene.GameAppId, sceneId)
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

func defaultDispatch(role, accountId, sceneId string) (*DispatchInfo, error) {
	game, err := chooseGameApp(role)
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
		err = setGameAppIdToSession(role, accountId, game.Uuid, sceneId)
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
		if err != nil {
			return nil, err
		}
		if session != nil && session.GameAppId != "" {
			return dispatchInfo(session.GameAppId)
		} else {
			return nil, nil
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

func dispatchScene(role string, sceneId string) (*scene_utils.Scene, error) {
	game, err := chooseGameApp(role)
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

func chooseGameApp(role string) (*game_utils.Game, error) {
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

func setGameAppIdToSession(role, accountId, gameAppId, sceneId string) error {
	session, err := session_utils.Find(accountId)
	if err != nil {
		return err
	}
	if session == nil {
		session, err = session_utils.Create(&session_utils.Session{
			AccountId: accountId,
			GameRole:  role,
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
