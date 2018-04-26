package agent_mgr

import (
	"goslib/redisdb"
	"strconv"
	"goslib/gen_server"
	"goslib/logger"
	"time"
	"strings"
)

type connectApp struct {
	Uuid string
	Name string
	Host string
	Port string
	Ccu  int
	CcuMax int
	Status int
}

type DispatchCache struct {
	app *connectApp
	activeAt int64
}

const (
	SERVER_STATUS_WORKING = iota
	SERVER_STATUS_MAINTAIN
)

const CONNECT_APP_IDS_KEY = "__CONNECT_APP_IDS_KEY"
const SERVER = "ConnectAppDispatcher"

func startAgentDispatcher() {
	SetupForTest()
	gen_server.Start(SERVER, new(Dispatcher))
}

func SetupForTest() {
	for i := 0; i < 10; i++  {
		num := strconv.Itoa(i)
		uuid := "fake_agent:" + num
		value := make(map[string]interface{})
		value["uuid"] = uuid
		value["name"] = "agent:" + num
		value["host"] = "127.0.0.1"
		value["port"] = "400" + num
		value["ccu"] = 0
		value["ccuMax"] = 100
		value["status"] = SERVER_STATUS_WORKING
		redisdb.Instance().HMSet(uuid, value)
		redisdb.Instance().SAdd(CONNECT_APP_IDS_KEY, uuid)
	}
}

func dispatchAgent(accountId string, groupId string) (appId string, host string, port string, err error) {
	result, err := gen_server.Call(SERVER, "Dispatch", accountId, groupId)
	if err != nil {
		logger.ERR("connectApp Dispatch failed: %v", err)
		return "", "", "", err
	}
	app := result.(*connectApp)
	if app == nil {
		return "", "", "", nil
	}

	appId = app.Uuid
	host = app.Host
	port = app.Port
	return
}

/*
   GenServer Callbacks
*/
type Dispatcher struct {
	apps []*connectApp
	mapApps map[string]*connectApp
	dispatchCache map[string]*DispatchCache
	appMapGroups map[string][]string
	groupMapApps map[string][]string

	printTimer *time.Timer
}

func (self *Dispatcher) startPrintTimer() {
	self.printTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(SERVER, "printStatus")
	})
}

func (self *Dispatcher) Init(args []interface{}) (err error) {
	self.LoadApps()
	self.dispatchCache = make(map[string]*DispatchCache)
	self.appMapGroups = make(map[string][]string)
	self.groupMapApps = make(map[string][]string)

	self.startPrintTimer()
	return nil
}

func (self *Dispatcher) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "printStatus" {
		logger.WARN("=============Load Balance===================")
		for _, app := range self.apps {
			logger.INFO(app.Uuid, " ccu: ", app.Ccu)
		}
		logger.WARN("=============App Groups===================")
		for appId, groupIds := range self.appMapGroups {
			logger.INFO(appId, " groupIds: ", strings.Join(groupIds, ","))
		}
		logger.WARN("=============Group Apps===================")
		for groupId, appIds := range self.groupMapApps {
			logger.INFO(groupId, " appIds: ", strings.Join(appIds, ","))
		}
		self.startPrintTimer()
	}
}

func (self *Dispatcher) HandleCall(args []interface{}) interface{} {
	handle := args[0].(string)
	if handle == "Dispatch" {
		accountId := args[1].(string)
		groupId := args[2].(string)
		return self.doDispatch(accountId, groupId)
	}
	return nil
}

func (self *Dispatcher) Terminate(reason string) (err error) {
	return nil
}

func (self *Dispatcher)LoadApps() {
	self.mapApps = make(map[string]*connectApp)
	ids, _ := redisdb.Instance().SMembers(CONNECT_APP_IDS_KEY).Result()
	self.apps = make([]*connectApp, 0)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		valueMap, _ := redisdb.Instance().HGetAll(id).Result()
		app := parseConnectApp(valueMap)
		self.apps = append(self.apps, app)
		self.mapApps[app.Uuid] = app
	}
	logger.INFO("idSize: ", len(ids), " appSize: ", len(self.apps))
}

func parseConnectApp(valueMap map[string]string) *connectApp {
	ccu, _ := strconv.Atoi(valueMap["ccu"])
	ccuMax, _ := strconv.Atoi(valueMap["ccuMax"])
	status, _ := strconv.Atoi(valueMap["status"])
	return &connectApp{
		valueMap["uuid"],
		valueMap["name"],
		valueMap["host"],
		valueMap["port"],
		ccu,
		ccuMax,
		status,
	}
}

/*
 * 关于连接服务的路由思考
 *	如果groupId不为空，优先将相同服玩家分配到相同代理，如果没有空间则寻找最空闲代理，建立分部
 *	如果groupId为空，直接分配到空闲服务器，没有空闲服务器就分配到负载最低的服务器
 */
func (self *Dispatcher) doDispatch(accountId string, groupId string) *connectApp {
	if cache, ok := self.dispatchCache[accountId]; ok {
		if self.matchDispatch(accountId, groupId, cache.app) {
			cache.activeAt = time.Now().Unix()
			return cache.app
		}
	}

	var dispatchApp *connectApp
	if groupId == "" {
		dispatchApp = self.dispatchByAccountId(accountId)
	} else {
		dispatchApp = self.dispatchByGroupId(accountId, groupId)
		self.appendAppIdToGroup(dispatchApp.Uuid, groupId)
		self.appendGroupIdToApp(dispatchApp.Uuid, groupId)
	}

	dispatchApp.Ccu++

	self.dispatchCache[accountId] = &DispatchCache{
		dispatchApp,
		time.Now().Unix(),
	}

	return dispatchApp
}

func (self *Dispatcher)appendAppIdToGroup(appId string, groupId string) {
	list, ok := self.groupMapApps[groupId]
	if !ok {
		list = []string{appId}
	} else {
		for _, aid := range list {
			if aid == appId {
				return
			}
		}
		list = append(list, appId)
	}
	self.groupMapApps[groupId] = list
}

func (self *Dispatcher)appendGroupIdToApp(appId string, groupId string) {
	list, ok := self.appMapGroups[appId]
	if !ok {
		list = []string{groupId}
	} else {
		for _, gid := range list {
			if gid == groupId {
				return
			}
		}
		list = append(list, groupId)
	}
	self.appMapGroups[appId] = list
}

func (self *Dispatcher)dispatchByAccountId(accountId string) *connectApp {
	var minPressureApp *connectApp
	for _, app := range self.apps {
		if app.Status != SERVER_STATUS_WORKING {
			continue
		}
		if app.Ccu < app.CcuMax {
			return app
		}
		minPressureApp = chooseLessPresure(minPressureApp, app, 0)
	}

	return minPressureApp
}

func (self *Dispatcher)dispatchByGroupId(accountId string, groupId string) *connectApp {
	appIds, ok := self.groupMapApps[groupId]
	var minPresureApp *connectApp
	var minGroupedPresureApp *connectApp

	// Dispatch to old group
	if ok {
		for _, appId := range appIds  {
			app := self.mapApps[appId]
			if app.Status != SERVER_STATUS_WORKING {
				continue
			}
			if app.Ccu < app.CcuMax {
				return app
			}
			minGroupedPresureApp = chooseLessPresure(minGroupedPresureApp, app, 0)
		}
	}

	// Dispatch to min presure app
	for _, app := range self.apps {
		if app.Status != SERVER_STATUS_WORKING {
			continue
		}
		minPresureApp = chooseLessPresure(minPresureApp, app, 0)
	}

	return chooseLessPresure(minPresureApp, minGroupedPresureApp, 0.3)
}

/*
 * Check targetApp is a valid agent for account
 *  1.has enough space
 *  2.is same group
 */
func (self *Dispatcher)matchDispatch(accountId string, groupId string, targetApp *connectApp) bool {
	if targetApp.Status != SERVER_STATUS_WORKING {
		return false
	}
	if groupId == ""{
		return self.matchDispatchByAccountId(accountId, targetApp)
	} else {
		return self.matchDispatchByGroupId(groupId, targetApp)
	}
}

func (self *Dispatcher)matchDispatchByAccountId(accountId string, targetApp *connectApp) bool {
	return targetApp.Ccu < targetApp.CcuMax
}

func (self *Dispatcher)matchDispatchByGroupId(groupId string, targetApp *connectApp) bool {
	appIds, ok := self.groupMapApps[groupId]
	if !ok {
		return false
	}

	for _, appId := range appIds  {
		app := self.mapApps[appId]
		if app.Uuid == targetApp.Uuid {
			return true
		}
	}

	return false
}

func chooseLessPresure(appA *connectApp, appB *connectApp, weightB float32) *connectApp {
	if appA == nil {
		return appB
	}

	if appB == nil {
		return appA
	}

	presureA := float32(appA.Ccu) / float32(appA.CcuMax)
	presureB := float32(appB.Ccu) / float32(appB.CcuMax)

	if presureA < presureB / (1 + weightB) {
		return appA
	} else {
		return appB
	}
}
