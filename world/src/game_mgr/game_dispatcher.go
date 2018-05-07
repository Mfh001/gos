package game_mgr

import (
	"github.com/kataras/iris/core/errors"
	"gosconf"
	. "goslib/game_utils"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/redisdb"
	. "goslib/scene_utils"
	"sync"
	"time"
)

const DISPATCH_SERVER = "GameAppDispatcher"

func startGameDispatcher() {
	gen_server.Start(DISPATCH_SERVER, new(Dispatcher))
}

type DispatchInfo struct {
	AppId   string
	AppHost string
	AppPort string
	SceneId string
}

type GameInfo struct {
	uuid     string
	host     string
	port     string
	ccu      int32
	activeAt int64
}

var gameInfos = &sync.Map{}

func dispatchGame(accountId string, serverId string, sceneId string) (*DispatchInfo, error) {
	result, err := gen_server.Call(DISPATCH_SERVER, "Dispatch", accountId, serverId, sceneId)
	if err != nil {
		logger.ERR("connectApp Dispatch failed: %v", err)
		return nil, err
	}
	info := result.(*DispatchInfo)
	return info, nil
}

func reportGameInfo(uuid string, host string, port string, ccu int32) {
	if gameInfo, ok := gameInfos.Load(uuid); ok {
		gameInfo.(*GameInfo).ccu = ccu
		gameInfo.(*GameInfo).activeAt = time.Now().Unix()
	} else {
		gameInfo := &GameInfo{
			uuid:     uuid,
			host:     host,
			port:     port,
			ccu:      ccu,
			activeAt: time.Now().Unix(),
		}
		gameInfos.Store(uuid, gameInfo)
	}
}

/*
   GenServer Callbacks
*/
type Dispatcher struct {
	defaultServerSceneConf *SceneConf

	apps map[string]*Game

	scenes map[string]*Scene

	printTimer *time.Timer
}

func (self *Dispatcher) startPrintTimer() {
	self.printTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(DISPATCH_SERVER, "printStatus")
	})
}

func (self *Dispatcher) startGameCheckTimer() {
	self.printTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(DISPATCH_SERVER, "gameCheck")
	})
}

func (self *Dispatcher) Init(args []interface{}) (err error) {
	self.apps = make(map[string]*Game)
	LoadGames(self.apps)
	for uuid, app := range self.apps {
		gameInfos.Store(uuid, &GameInfo{
			uuid:     uuid,
			host:     app.Host,
			port:     app.Port,
			ccu:      app.Ccu,
			activeAt: time.Now().Unix(),
		})
	}

	self.scenes = make(map[string]*Scene)
	LoadScenes(self.scenes)
	for _, scene := range self.scenes {
		if _, ok := self.apps[scene.Uuid]; !ok {
			self.dispatchScene(scene)
		}
	}

	self.InitDefaultServerScene()

	self.startPrintTimer()
	self.startGameCheckTimer()
	return nil
}

func (self *Dispatcher) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "printStatus" {
		for _, app := range self.apps {
			activeAt := time.Unix(app.ActiveAt, 0)
			logger.INFO("Game uuid: ", app.Uuid, " address: ", app.Host, ":", app.Port, " ccu: ", app.Ccu, " activeAt: ", activeAt)
		}
		//logger.WARN("=============App Served Scenes===================")
		//for appId, groupIds := range self.appMapScenes {
		//	logger.INFO(appId, " groupIds: ", strings.Join(groupIds, ","))
		//}
		self.startPrintTimer()
	} else if handle == "gameCheck" {
		var needDelIds = make([]string, 0)
		gameInfos.Range(func(key, value interface{}) bool {
			gameInfo := value.(*GameInfo)
			now := time.Now().Unix()
			if isGameAlive(now, gameInfo.activeAt) {
				if game, ok := self.apps[gameInfo.uuid]; ok {
					game.Ccu = gameInfo.ccu
					game.ActiveAt = now
				} else {
					logger.WARN("addGame: ", gameInfo.uuid)
					self.addGame(gameInfo)
				}
			} else {
				logger.WARN("delGame: ", gameInfo.uuid)
				needDelIds = append(needDelIds, gameInfo.uuid)
				self.delGame(gameInfo.uuid)
			}
			return true
		})
		for _, needDelId := range needDelIds {
			gameInfos.Delete(needDelId)
		}
		self.startGameCheckTimer()
	}
}

func isGameAlive(now int64, activeAt int64) bool {
	return activeAt+gosconf.SERVICE_DEAD_DURATION > now
}

func (self *Dispatcher) addGame(info *GameInfo) {
	app := &Game{
		Uuid:     info.uuid,
		Host:     info.host,
		Port:     info.port,
		Ccu:      0,
		CcuMax:   gosconf.GAME_CCU_MAX,
		ActiveAt: time.Now().Unix(),
	}
	self.apps[app.Uuid] = app
	redisdb.ServiceInstance().SAdd(gosconf.RK_GAME_APP_IDS, app.Uuid)
	app.Save()
}

func (self *Dispatcher) delGame(uuid string) {
	if app, ok := self.apps[uuid]; ok {
		redisdb.ServiceInstance().SRem(gosconf.RK_GAME_APP_IDS, app.Uuid)
		app.Del()
	}
	delete(self.apps, uuid)
	for _, scene := range self.scenes {
		if scene.GameAppId == uuid {
			self.dispatchScene(scene)
		}
	}
}

func (self *Dispatcher) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "Dispatch" {
		accountId := args[1].(string)
		serverId := args[2].(string)
		sceneId := args[3].(string)
		return self.doDispatch(accountId, serverId, sceneId)
	}
	return nil, nil
}

func (self *Dispatcher) Terminate(reason string) (err error) {
	return nil
}

func (self *Dispatcher) InitDefaultServerScene() {
	sceneConf, err := FindSceneConf(gosconf.RK_DEFAULT_SERVER_SCENE_CONF_ID)
	if err != nil {
		logger.ERR("Init Default Server Scene failed!")
	}
	self.defaultServerSceneConf = sceneConf
}

/*
 * 关于游戏服务的路由思考
 *  如果sceneId为空，根据serverId将玩家分配到对应游戏服务
 *	如果sceneId不为空，根据serverId和sceneId将玩家分配到对应游戏服务
 */
func (self *Dispatcher) doDispatch(accountId string, serverId string, sceneId string) (*DispatchInfo, error) {
	var dispatchApp *Game
	var dispatchScene *Scene
	var err error
	logger.INFO("doDispatch accountId: ", accountId, " serverId: ", serverId, " sceneId: ", sceneId)
	if sceneId == "" {
		dispatchApp, dispatchScene, err = self.dispatchByServerId(serverId)
	} else {
		dispatchApp, dispatchScene, err = self.dispatchBySceneId(sceneId)
	}

	if err != nil {
		logger.ERR("Dispatch Game Failed: ", err)
		return nil, err
	}

	dispatchApp.Ccu++

	return &DispatchInfo{
		AppId:   dispatchApp.Uuid,
		AppHost: dispatchApp.Host,
		AppPort: dispatchApp.Port,
		SceneId: dispatchScene.Uuid,
	}, nil
}

func (self *Dispatcher) dispatchByServerId(serverId string) (*Game, *Scene, error) {
	// Lookup scene
	sceneIns, ok := self.scenes[serverId]

	if !ok {
		sceneIns, err := CreateDefaultServerScene(serverId, self.defaultServerSceneConf)
		if err != nil {
			return nil, nil, err
		}
		self.scenes[serverId] = sceneIns
		return self.dispatchedInfo(sceneIns)
	}

	logger.INFO("sceneIns: ", sceneIns.Uuid)
	return self.dispatchedInfo(sceneIns)
}

func (self *Dispatcher) dispatchBySceneId(sceneId string) (*Game, *Scene, error) {
	// Lookup scene
	sceneIns, ok := self.scenes[sceneId]
	if !ok {
		return nil, nil, nil
	}

	return self.dispatchedInfo(sceneIns)
}

func (self *Dispatcher) dispatchedInfo(sceneIns *Scene) (*Game, *Scene, error) {
	err := self.makeSureSceneDispatched(sceneIns)
	if err != nil {
		return nil, nil, err
	}

	gameApp, ok := self.apps[sceneIns.GameAppId]
	if !ok {
		return nil, nil, nil
	}

	return gameApp, sceneIns, nil
}

func (self *Dispatcher) makeSureSceneDispatched(scene *Scene) error {
	// Dispath scene
	_, ok := self.apps[scene.GameAppId]
	if scene.GameAppId == "" || !ok {
		err := self.dispatchScene(scene)
		if err != nil {
			logger.ERR("makeSureSceneDispatched: ", scene.Uuid, " failed: ", err)
			return err
		}
	}
	return nil
}

// Dispatch scene to specific gameApp
func (self *Dispatcher) dispatchScene(scene *Scene) error {
	var minPresureApp *Game

	// Dispatch to min presure app
	for _, app := range self.apps {
		minPresureApp = chooseLessPresure(minPresureApp, app, 0)
	}

	if minPresureApp == nil {
		return errors.New("No working Game")
	}

	minPresureApp.Ccu = minPresureApp.Ccu + int32(scene.CcuMax)
	scene.GameAppId = minPresureApp.Uuid
	//sceneCell.DeploySceneToGameApp(scene)

	params := make(map[string]interface{})
	params["gameAppId"] = scene.GameAppId
	redisdb.ServiceInstance().HMSet(scene.Uuid, params)

	return nil
}

func chooseLessPresure(appA *Game, appB *Game, weightB float32) *Game {
	if appA == nil {
		return appB
	}

	if appB == nil {
		return appA
	}

	presureA := float32(appA.Ccu) / float32(appA.CcuMax)
	presureB := float32(appB.Ccu) / float32(appB.CcuMax)

	if presureA < presureB/(1+weightB) {
		return appA
	} else {
		return appB
	}
}
