package agent_mgr

import (
	"goslib/gen_server"
	"goslib/logger"
	"time"
	"gosconf"
	"github.com/kataras/iris/core/errors"
	"sync"
)

type connectApp struct {
	Uuid string
	Host string
	Port string
	Ccu  int32
	CcuMax int32
	ActiveAt int64
}

type DispatchCache struct {
	app *connectApp
	activeAt int64
}

const SERVER = "ConnectAppDispatcher"
var agentInfos = &sync.Map{}
type AgentInfo struct {
	uuid string
	host string
	port string
	ccu int32
	activeAt int64
}

func startAgentDispatcher() {
	gen_server.Start(SERVER, new(Dispatcher))
}

func handleReportAgentInfo(uuid string, host string, port string, ccu int32) {
	if agentInfo, ok := agentInfos.Load(uuid); ok {
		agentInfo.(*AgentInfo).ccu = ccu
		agentInfo.(*AgentInfo).activeAt = time.Now().Unix()
	} else {
		agentInfo = &AgentInfo{
			uuid: uuid,
			host: host,
			port: port,
			ccu: ccu,
			activeAt: time.Now().Unix(),
		}
		agentInfos.Store(uuid, agentInfo)
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
	apps          map[string]*connectApp
	dispatchCache map[string]*DispatchCache
	appMapGroups  map[string][]string
	groupMapApps  map[string][]string

	printTimer *time.Timer
	agentCheckTimer *time.Timer
}

func (self *Dispatcher) startPrintTimer() {
	self.printTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(SERVER, "printStatus")
	})
	self.agentCheckTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(SERVER, "agentCheck")
	})
}

func (self *Dispatcher) startAgentCheckTimer() {
	self.agentCheckTimer = time.AfterFunc(5*time.Second, func() {
		gen_server.Cast(SERVER, "agentCheck")
	})
}

func (self *Dispatcher) Init(args []interface{}) (err error) {
	self.apps = make(map[string]*connectApp)
	self.dispatchCache = make(map[string]*DispatchCache)
	self.appMapGroups = make(map[string][]string)
	self.groupMapApps = make(map[string][]string)

	self.startPrintTimer()
	self.startAgentCheckTimer()
	return nil
}

func (self *Dispatcher) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "printStatus" {
		for _, app := range self.apps {
			activeAt := time.Unix(app.ActiveAt, 0)
			logger.INFO("Agent uuid: ", app.Uuid, " address: ", app.Host, ":", app.Port, " ccu: ", app.Ccu, " activeAt: ", activeAt)
		}
		//logger.WARN("=============App Groups===================")
		//for appId, groupIds := range self.appMapGroups {
		//	logger.INFO(appId, " groupIds: ", strings.Join(groupIds, ","))
		//}
		//logger.WARN("=============Group Apps===================")
		//for groupId, appIds := range self.groupMapApps {
		//	logger.INFO(groupId, " appIds: ", strings.Join(appIds, ","))
		//}
		self.startPrintTimer()
	} else if handle == "agentCheck" {
		now := time.Now().Unix()
		var needDelIds = make([]string, 0)
		agentInfos.Range(func(key, value interface{}) bool {
			agentInfo := value.(*AgentInfo)
			if isAgentAlive(now, agentInfo.activeAt) {
				if agent, ok := self.apps[agentInfo.uuid]; ok {
					agent.Ccu = agentInfo.ccu
					agent.ActiveAt = now
				} else {
					logger.WARN("addAgent: ", agentInfo.uuid)
					self.addAgent(agentInfo)
				}
			} else {
				logger.WARN("delAgent: ", agentInfo.uuid)
				needDelIds = append(needDelIds, agentInfo.uuid)
				self.delAgent(agentInfo.uuid)
			}
			return true
		})
		for _, needDelId := range needDelIds {
			agentInfos.Delete(needDelId)
		}
		self.startAgentCheckTimer()
	}
}

func isAgentAlive(now int64, activeAt int64) bool {
	return activeAt + gosconf.SERVICE_DEAD_DURATION > now
}

func (self *Dispatcher) addAgent(info *AgentInfo)  {
	agent := &connectApp{
		Uuid: info.uuid,
		Host: info.host,
		Port: info.port,
		Ccu: 0,
		CcuMax: gosconf.AGENT_CCU_MAX,
		ActiveAt: time.Now().Unix(),
	}
	self.apps[agent.Uuid] = agent
}

func (self *Dispatcher) delAgent(uuid string)  {
	delete(self.apps, uuid)
}

func (self *Dispatcher) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "Dispatch" {
		accountId := args[1].(string)
		groupId := args[2].(string)
		return self.doDispatch(accountId, groupId)
	}
	return nil, nil
}

func (self *Dispatcher) Terminate(reason string) (err error) {
	return nil
}

/*
 * 关于连接服务的路由思考
 *	如果groupId不为空，优先将相同服玩家分配到相同代理，如果没有空间则寻找最空闲代理，建立分部
 *	如果groupId为空，直接分配到空闲服务器，没有空闲服务器就分配到负载最低的服务器
 */
func (self *Dispatcher) doDispatch(accountId string, groupId string) (*connectApp, error) {
	if cache, ok := self.dispatchCache[accountId]; ok {
		if cache == nil {
			delete(self.dispatchCache, accountId)
		} else {
			if self.matchDispatch(accountId, groupId, cache.app) {
				cache.activeAt = time.Now().Unix()
				return cache.app, nil
			}
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

	if dispatchApp == nil {
		return nil, errors.New("No working agent found!")
	}

	dispatchApp.Ccu++

	self.dispatchCache[accountId] = &DispatchCache{
		dispatchApp,
		time.Now().Unix(),
	}

	return dispatchApp, nil
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
			app := self.apps[appId]
			if app.Ccu < app.CcuMax {
				return app
			}
			minGroupedPresureApp = chooseLessPresure(minGroupedPresureApp, app, 0)
		}
	}

	// Dispatch to min presure app
	for _, app := range self.apps {
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
		app := self.apps[appId]
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
