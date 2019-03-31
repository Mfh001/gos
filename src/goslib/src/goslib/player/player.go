package player

import (
	"errors"
	"fmt"
	"gen/api/pt"
	"gen/db"
	"gosconf"
	"goslib/api"
	"goslib/broadcast"
	"goslib/game_server/interfaces"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/player_data"
	"goslib/routes"
	"goslib/scene_mgr"
	"goslib/session_utils"
	"runtime"
	"time"
)

type Player struct {
	PlayerId     string
	ProcessDict  map[string]interface{}
	Data         *db.PlayerData
	Agents       map[string]interfaces.AgentBehavior
	processed    int
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
}

type RPCReply struct {
	EncodeMethod string
	Response     interface{}
}

const EXPIRE_DURATION = 1800

var BroadcastHandler func(*Player, *broadcast.BroadcastMsg) = nil
var CurrentGameAppId string

func Connected(accountId string, agentId string, agent interfaces.AgentBehavior) error {
	return CastPlayer(accountId, "connected", agentId, agent)
}

func Disconnected(accountId, agentId string) error {
	return CastPlayer(accountId, "disconnected", agentId)
}

func HandleRequest(accountId, agentId string, requestData []byte) error {
	return CastPlayer(accountId, "handleRequest", agentId, requestData)
}

func HandleRPCCall(accountId string, requestData []byte) ([]byte, error) {
	handler, params, err := api.ParseRequestDataForHander(requestData)
	if err != nil {
		return nil, err
	}
	result, err := CallPlayer(accountId, "handleRPCCall", handler, params)
	if err != nil {
		logger.ERR("HandleRPCCall failed: ", err)
		return nil, err
	}
	reply := result.(*RPCReply)
	return EncodeResponseData(reply.EncodeMethod, reply.Response)
}

func CallPlayer(accountId string, args ...interface{}) (interface{}, error) {
	if !gen_server.Exists(accountId) {
		err := StartPlayer(accountId)
		if err != nil {
			return nil, err
		}
	}
	return gen_server.Call(accountId, args...)
}

func CastPlayer(accountId string, args ...interface{}) error {
	if !gen_server.Exists(accountId) {
		err := StartPlayer(accountId)
		if err != nil {
			return err
		}
	}
	gen_server.Cast(accountId, args...)
	return nil
}

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	name := args[0].(string)
	fmt.Println("Player: ", name, " started!")
	self.PlayerId = name
	self.ProcessDict = make(map[string]interface{})
	self.Data, err = player_data.Load(self.PlayerId)
	if err != nil {
		logger.ERR("Start player failed, cannot load PlayerData: ", err)
		return err
	}
	self.lastActive = time.Now().Unix()
	self.Agents = make(map[string]interfaces.AgentBehavior)
	self.startActiveCheck()
	self.startPersistTimer()

	session, err := session_utils.Find(self.PlayerId)
	if err != nil {
		logger.ERR("Player lookup session failed: ", self.PlayerId, " err: ", err)
	} else {
		scene_mgr.TryLoadScene(session.SceneId)
	}

	return nil
}

func (self *Player) startPersistTimer() {
	self.persistTimer = time.AfterFunc(300*time.Second, func() {
		gen_server.Cast(self.PlayerId, "PersistData")
	})
}

func (self *Player) HandleCast(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "handleRequest" {
		_ = self.handleRequest(args[1].(string), args[2].([]byte))
	} else if method_name == "handleRPCCast" {
		self.handleRPCCast(args[1].([]byte))
	} else if method_name == "handleWrap" {
		self.handleWrap(args[1].(func(player *Player) interface{}))
	} else if method_name == "handleAsyncWrap" {
		self.handleAsyncWrap(args[0].(func()))
	} else if method_name == "PersistData" {
		//err := self.Store.Persist([]string{"models"})
		err := player_data.Persist(self.PlayerId, self.Data)
		if err != nil {
			logger.ERR("PersistData failed: ", err)
		}
		self.startPersistTimer()
	} else if method_name == "removeConn" {
		//self.Conn = nil
	} else if method_name == "broadcast" {
		self.handleBroadcast(args[1].(*broadcast.BroadcastMsg))
	} else if method_name == "connected" {
		agentId := args[1].(string)
		agent := args[2].(interfaces.AgentBehavior)
		self.Agents[agentId] = agent
	} else if method_name == "disconnected" {
		agentId := args[1].(string)
		delete(self.Agents, agentId)
	}
}

func (self *Player) HandleCall(args []interface{}) (interface{}, error) {
	methodName := args[0].(string)
	if methodName == "handleWrap" {
		return self.handleWrap(args[1].(func(player *Player) interface{})), nil
	} else if methodName == "handleRPCCall" {
		return self.handleRPCCall(args[1].(routes.Handler), args[2])
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
	return
}

func (self *Player) startActiveCheck() {
	if (self.lastActive + EXPIRE_DURATION) < time.Now().Unix() {
		err := gen_server.Stop(self.PlayerId, "Shutdown inActive player!")
		logger.ERR("Stop gen_server failed: ", err)
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

func (self *Player) SendData(agentId, encode_method string, msg interface{}) error {
	return self.sendDataToStream(agentId, encode_method, msg)
}

func (self *Player) handleRequest(agentId string, data []byte) error {
	self.lastActive = time.Now().Unix()
	if !gosconf.IS_DEBUG {
		defer func() {
			if x := recover(); x != nil {
				logger.ERR("caught panic in player handleRequest(): ", x)
			}
		}()
	}

	handler, params, err := api.ParseRequestDataForHander(data)
	if err != nil {
		logger.ERR("ParseRequestDataForHander failed: ", err)
		return self.sendDataToStream(agentId, pt.PT_Fail, &pt.Fail{Fail: "error_route_not_found"})
	} else {
		encode_method, msg := self.processRequest(handler, params)
		return self.sendDataToStream(agentId, encode_method, msg)
	}
}

func (self *Player) handleRPCCall(handler routes.Handler, params interface{}) (*RPCReply, error) {
	encode_method, response := handler(self, params)
	return &RPCReply{EncodeMethod: encode_method, Response: response}, nil
}

func (self *Player) handleRPCCast(data []byte) {
	handler, params, err := api.ParseRequestDataForHander(data)
	if err != nil {
		logger.ERR("handleRPCCast failed: ", err)
		return
	}
	self.processRequest(handler, params)
}

func EncodeResponseData(encode_method string, response interface{}) ([]byte, error) {
	writer, err := api.Encode(encode_method, response)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	return writer.GetSendData()
}

func (self *Player) processRequest(handler routes.Handler, params interface{}) (string, interface{}) {
	encode_method, response := handler(self, params)
	self.processed++
	logger.INFO("Processed: ", self.processed, " Response Data: ", response)
	return encode_method, response
}

func (self *Player) handleWrap(fun func(ctx *Player) interface{}) interface{} {
	self.lastActive = time.Now().Unix()
	return fun(self)
}

func (self *Player) handleAsyncWrap(fun func()) {
	self.lastActive = time.Now().Unix()
	fun()
}

func (self *Player) handleBroadcast(msg *broadcast.BroadcastMsg) {
	if BroadcastHandler != nil {
		BroadcastHandler(self, msg)
	}
}

/*
   IPC Methods
*/

func (self *Player) Wrap(targetPlayerId string, fun func(ctx *Player) interface{}) (interface{}, error) {
	if self.PlayerId == targetPlayerId {
		return self.handleWrap(fun), nil
	} else {
		return CallPlayer(targetPlayerId, "handleWrap", fun)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, fun func()) {
	if self.PlayerId == targetPlayerId {
		self.handleAsyncWrap(fun)
	} else {
		err := CastPlayer(targetPlayerId, "HandleAsyncWrap", fun)
		if err != nil {
			logger.ERR("HandleAsyncWrap failed: ", err)
		}
	}
}

func (self *Player) sendDataToStream(agentId, encode_method string, msg interface{}) error {
	agent, ok := self.Agents[agentId]
	if !ok {
		for key := range self.Agents {
			logger.INFO("have agentId: ", key)
		}
		errMsg := fmt.Sprintf("sendDataToStream failed, agent not exists: %s", agentId)
		return errors.New(errMsg)
	}
	writer, err := api.Encode(encode_method, msg)
	if err != nil {
		logger.ERR("encode data failed: ", err)
		return err
	}
	data, err := writer.GetSendData()
	if err != nil {
		logger.ERR("encrypt data failed: ", err)
		return err
	}
	return sendToClient(data, agent)
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
