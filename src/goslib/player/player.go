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
package player

import (
	"container/heap"
	"errors"
	"fmt"
	"github.com/mafei198/gos/goslib/api"
	"github.com/mafei198/gos/goslib/broadcast"
	"github.com/mafei198/gos/goslib/game_server/interfaces"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/gen/db"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player_data"
	"github.com/mafei198/gos/goslib/routes"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/goslib/utils"
	"runtime"
	"time"
)

type Player struct {
	PlayerId     string
	Data         *db.PlayerData
	DataCheckSum string
	Agents       map[string]interfaces.AgentBehavior
	processed    int
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
	SceneId      string

	TimerTasks       utils.PriorityQueue
	TimerTaskItems   map[string]*utils.Item
	TimerTasksInited bool
	TimerTasksTimer  *time.Timer
}

type RPCReply struct {
	Response interface{}
}

type LifeCycle interface {
	OnInit(ctx *Player, args []interface{})
	OnConnected(ctx *Player, agentId string)
	OnDisConnected(ctx *Player, agentId string)
	OnTerminate(ctx *Player, reason string)
}

const EXPIRE_DURATION = 1800

var IsTesting = false

var lifeCycle LifeCycle
var BroadcastHandler func(*Player, *broadcast.BroadcastMsg) = nil
var CurrentGameAppId string

func RegistLifeCycle(handler LifeCycle) {
	lifeCycle = handler
}

func Connected(accountId string, agentId string, agent interfaces.AgentBehavior) error {
	return CastPlayer(accountId, &ConnectedParams{agentId, agent})
}

func Disconnected(accountId, agentId, connId string) {
	gen_server.Cast(accountId, &DisConnectedParams{agentId, connId})
}

func HandleRequest(accountId, agentId string, requestData []byte) error {
	return CastPlayer(accountId, &RequestParams{agentId, requestData})
}

func HandleRPCCall(accountId string, category int32, requestData []byte) ([]byte, error) {
	switch category {
	case gosconf.RPC_CALL_NORMAL:
		reqId, handler, params, err := api.ParseRequestDataForHander(requestData)
		if err != nil {
			return nil, err
		}
		result, err := CallPlayer(accountId, &RpcCallParams{handler, params})
		if err != nil {
			logger.ERR("HandleRPCCall failed: ", err)
			return nil, err
		}
		reply := result.(*RPCReply)
		return EncodeResponseData(reqId, reply.Response)
	case gosconf.RPC_CALL_PROXY_DATA:
		err := CastPlayer(accountId, &ProxyData{
			AgentId: accountId,
			Data:    requestData,
		})
		return nil, err
	}
	return nil, nil
}

func SendData(accountId, agentId string, msg interface{}) bool {
	if gen_server.Exists(accountId) {
		gen_server.Cast(accountId, &SendDataParam{agentId, msg})
		return true
	}
	return false
}

func GetCtx(accountId string) (*db.PlayerData, error) {
	result, err := CallPlayer(accountId, &GetCtxParams{})
	if err != nil {
		return nil, err
	}
	return result.(*db.PlayerData), nil
}

func CallPlayer(accountId string, msg interface{}) (interface{}, error) {
	if !gen_server.Exists(accountId) {
		err := StartPlayer(accountId)
		if err != nil {
			return nil, err
		}
	}
	return gen_server.Call(accountId, msg)
}

func CastPlayer(accountId string, msg interface{}) error {
	if !gen_server.Exists(accountId) {
		err := StartPlayer(accountId)
		if err != nil {
			return err
		}
	}
	gen_server.Cast(accountId, msg)
	return nil
}

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	accountId := args[0].(string)
	sceneId := args[1].(string)
	fmt.Println("Player: ", accountId, " started!")
	self.PlayerId = accountId
	self.SceneId = sceneId
	if !IsTesting {
		self.DataCheckSum, self.Data, err = player_data.Load(self.PlayerId)
		if err != nil {
			logger.ERR("Start player failed, cannot load PlayerData: ", err)
			return err
		}
	}
	self.lastActive = time.Now().Unix()
	self.Agents = make(map[string]interfaces.AgentBehavior)
	self.InitTimerTask()

	if lifeCycle != nil {
		lifeCycle.OnInit(self, args)
	}

	self.startActiveCheck()
	self.startPersistTimer()
	session_utils.Active(self.PlayerId)

	return nil
}

func (self *Player) InitTimerTask() {
	self.TimerTasks = make(utils.PriorityQueue, 0)
	self.TimerTaskItems = map[string]*utils.Item{}
	self.TimerTasksInited = false
	heap.Init(&self.TimerTasks)
}

func (self *Player) startPersistTimer() {
	self.persistTimer = time.AfterFunc(300*time.Second, func() {
		gen_server.Cast(self.PlayerId, &PersistData{})
	})
}

func (self *Player) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *RequestParams:
		_ = self.handleRequest(params)
		break
	case *RPCCastParams:
		self.handleRPCCast(params)
		break
	case *WrapParams:
		self.handleWrap(params)
		break
	case *AsyncWrapParams:
		self.handleAsyncWrap(params)
		break
	case *PersistData:
		self.handlePersistData()
		break
	case *BroadcastParams:
		self.handleBroadcast(params)
		break
	case *ConnectedParams:
		self.handleConnected(params)
		break
	case *DisConnectedParams:
		self.handleDisconnected(params)
		break
	case *SendDataParam:
		_ = self.SendData(params.agentId, params.msg)
		break
	case *ActiveSleepParams:
		self.lastActive = time.Now().Unix()
		self.startActiveCheck()
		break
	case *ProxyData:
		self.sendRawToClient(params.AgentId, params.Data)
		break
	case *CheckTimerTaskParams:
		self.CheckTimerTask()
		break
	}
}

type ProxyData struct {
	AgentId string
	Data    []byte
}

type GetCtxParams struct{}

func (self *Player) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *WrapParams:
		return self.handleWrap(params), nil
	case *RpcCallParams:
		return self.handleRPCCall(params)
	case *GetCtxParams:
		return self.Data, nil
	}
	return nil, nil
}

func (self *Player) Terminate(reason string) (err error) {
	fmt.Println("callback Termiante!")
	self.activeTimer.Stop()
	self.persistTimer.Stop()
	err = player_data.Return(self.PlayerId, self.Data)
	if err != nil {
		logger.ERR("Persist data failed: ", err)
		return
	}
	if lifeCycle != nil {
		lifeCycle.OnTerminate(self, reason)
	}
	session_utils.Active(self.PlayerId)
	return
}

type ActiveSleepParams struct{}

func (self *Player) startActiveCheck() {
	if (self.lastActive + EXPIRE_DURATION) < time.Now().Unix() {
		SleepPlayer(self.PlayerId)
	} else {
		self.activeTimer = time.AfterFunc(10*time.Second, self.startActiveCheck)
	}
}

/*
   IPC Methods
*/

func (self *Player) SystemInfo() int {
	return runtime.NumCPU()
}

type SendDataParam struct {
	agentId string
	msg     interface{}
}

func (self *Player) SendData(agentId string, msg interface{}) error {
	return self.sendDataToStream(agentId, 0, msg)
}

type RequestParams struct {
	agentId string
	data    []byte
}

func (self *Player) handleRequest(params *RequestParams) error {
	self.lastActive = time.Now().Unix()

	reqId, handler, args, err := api.ParseRequestDataForHander(params.data)
	if err != nil {
		logger.ERR("ParseRequestDataForHander failed: ", err)
		return self.sendDataToStream(params.agentId, reqId, &pt.Fail{Fail: "error_route_not_found"})
	} else {
		msg := self.processRequest(handler, args)
		return self.sendDataToStream(params.agentId, reqId, msg)
	}
}

type RpcCallParams struct {
	Handler routes.Handler
	Params  interface{}
}

func (self *Player) handleRPCCall(params *RpcCallParams) (*RPCReply, error) {
	response := params.Handler(self, params.Params)
	return &RPCReply{Response: response}, nil
}

type RPCCastParams struct{ data []byte }

func (self *Player) handleRPCCast(params *RPCCastParams) {
	_, handler, args, err := api.ParseRequestDataForHander(params.data)
	if err != nil {
		logger.ERR("handleRPCCast failed: ", err)
		return
	}
	self.processRequest(handler, args)
}

func EncodeResponseData(reqId int32, response interface{}) ([]byte, error) {
	writer, err := api.Encode(response)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	return writer.GetSendData(reqId)
}

func (self *Player) processRequest(handler routes.Handler, params interface{}) interface{} {
	response := handler(self, params)
	self.processed++
	return response
}

type WrapParams struct{ Fun func(ctx *Player) interface{} }

func (self *Player) handleWrap(params *WrapParams) interface{} {
	self.lastActive = time.Now().Unix()
	return params.Fun(self)
}

type AsyncWrapParams struct{ fun func(ctx *Player) }

func (self *Player) handleAsyncWrap(params *AsyncWrapParams) {
	self.lastActive = time.Now().Unix()
	params.fun(self)
}

type PersistData struct{}

func (self *Player) handlePersistData() {
	checkSum, err := player_data.Persist(self.PlayerId, self.DataCheckSum, self.Data)
	if err != nil {
		logger.ERR("PersistData failed: ", err)
	}
	self.DataCheckSum = checkSum
	session_utils.Active(self.PlayerId)
	self.startPersistTimer()
}

type BroadcastParams struct{ msg *broadcast.BroadcastMsg }

func (self *Player) handleBroadcast(params *BroadcastParams) {
	if BroadcastHandler != nil {
		BroadcastHandler(self, params.msg)
	}
}

type ConnectedParams struct {
	agentId string
	agent   interfaces.AgentBehavior
}

func (self *Player) handleConnected(params *ConnectedParams) {
	self.Agents[params.agentId] = params.agent
	if lifeCycle != nil {
		lifeCycle.OnConnected(self, params.agentId)
	}
}

type DisConnectedParams struct {
	agentId string
	connId  string
}

func (self *Player) handleDisconnected(params *DisConnectedParams) {
	agent, ok := self.Agents[params.agentId]
	if ok && agent.GetUuid() == params.connId {
		if lifeCycle != nil {
			lifeCycle.OnDisConnected(self, params.agentId)
		}
		delete(self.Agents, params.agentId)
	}
}

/*
   IPC Methods
*/

func (self *Player) EnterScene(agentId, sceneId string) error {
	session, err := session_utils.Find(agentId)
	if err != nil {
		return err
	}
	session.SceneId = sceneId
	if err := session.Save(); err != nil {
		return err
	}
	if agent, ok := self.Agents[agentId]; ok {
		return agent.ConnectScene(session)
	}
	return nil
}

func (self *Player) Wrap(targetPlayerId string, params *WrapParams) (interface{}, error) {
	if self.PlayerId == targetPlayerId {
		return self.handleWrap(params), nil
	} else {
		return CallPlayer(targetPlayerId, params)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, params *AsyncWrapParams) {
	if self.PlayerId == targetPlayerId {
		self.handleAsyncWrap(params)
	} else {
		err := CastPlayer(targetPlayerId, params)
		if err != nil {
			logger.ERR("HandleAsyncWrap failed: ", err)
		}
	}
}

type TimerTaskHandler func(ctx *Player)

func (self *Player) AddTimerTask(id string, runAt int, handler TimerTaskHandler) {
	oldHead := self.TimerTasks.HeadItem()
	item, ok := self.TimerTaskItems[id]
	if !ok {
		item = &utils.Item{
			Id:       id,
			Value:    handler,
			Priority: runAt,
		}
		self.TimerTaskItems[id] = item
		heap.Push(&self.TimerTasks, item)
	} else {
		self.TimerTasks.Update(item, handler, runAt)
	}
	newHead := self.TimerTasks.HeadItem()
	if oldHead == nil || newHead.Priority < oldHead.Priority {
		self.CheckTimerTask()
	}
}

func (self *Player) UpdateTimerRunAt(id string, runAt int) {
	item, ok := self.TimerTaskItems[id]
	if ok {
		oldHead := self.TimerTasks.HeadItem()
		self.TimerTasks.Update(item, item.Value.(TimerTaskHandler), runAt)
		newHead := self.TimerTasks.HeadItem()
		now := int(time.Now().Unix())
		if oldHead == nil || newHead.Priority < oldHead.Priority || newHead.Priority <= now {
			self.CheckTimerTask()
		}
	}
}

func (self *Player) DelTimerTask(id string) {
	if item, ok := self.TimerTaskItems[id]; ok {
		self.TimerTasks.Remove(item)
	}
}

type CheckTimerTaskParams struct{}

func (self *Player) CheckTimerTask() {
	item := self.TimerTasks.HeadItem()
	if item == nil {
		return
	}
	now := int(time.Now().Unix())
	if item.Priority > now {
		self.ScheduleTimerTask(item)
		return
	}

	handler := item.Value.(TimerTaskHandler)
	heap.Pop(&self.TimerTasks)
	self.TimerTasksTimer = nil
	delete(self.TimerTaskItems, item.Id)
	self.ScheduleTimerTask(self.TimerTasks.HeadItem())
	handler(self)
}

func (self *Player) ScheduleTimerTask(item *utils.Item) {
	if item == nil {
		return
	}
	now := int(time.Now().Unix())
	if self.TimerTasksTimer != nil {
		self.TimerTasksTimer.Stop()
	}

	if item.Priority > now {
		self.TimerTasksTimer = time.AfterFunc(time.Duration(item.Priority-now)*time.Second, func() {
			gen_server.Cast(self.PlayerId, &CheckTimerTaskParams{})
		})
	} else {
		gen_server.Cast(self.PlayerId, &CheckTimerTaskParams{})
	}
}

func (self *Player) sendDataToStream(agentId string, reqId int32, msg interface{}) error {
	logger.INFO("Response agentId: ", agentId, " processed: ", self.processed, " reqId: ", reqId, " params: ", utils.StructToStr(msg))
	agent, ok := self.Agents[agentId]
	if !ok {
		for key := range self.Agents {
			logger.INFO("have agentId: ", key)
		}
		errMsg := fmt.Sprintf("sendDataToStream failed, agent not exists: %s", agentId)
		return errors.New(errMsg)
	}
	writer, err := api.Encode(msg)
	if err != nil {
		logger.ERR("encode data failed: ", err)
		return err
	}
	data, err := writer.GetSendData(reqId)
	if err != nil {
		logger.ERR("encrypt data failed: ", err)
		return err
	}
	return sendToClient(data, agent)
}

func (self *Player) sendRawToClient(agentId string, data []byte) {
	logger.INFO("proxy data: ", agentId)
	agent := self.Agents[agentId]
	sendToClient(data, agent)
}

func sendToClient(data []byte, agent interfaces.AgentBehavior) error {
	if agent != nil {
		err := agent.SendMessage(data)
		if err != nil {
			logger.ERR("sendToClient failed: ", err)
			return err
		} else {
			return nil
		}
	} else {
		errMsg := "sendToClient failed, connectAppId is nil"
		logger.WARN(errMsg)
		return errors.New(errMsg)
	}
}
