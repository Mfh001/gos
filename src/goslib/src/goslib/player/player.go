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
	return CastPlayer(accountId, &ConnectedParams{agentId, agent})
}

func Disconnected(accountId, agentId string) error {
	return CastPlayer(accountId, &DisConnectedParams{agentId})
}

func HandleRequest(accountId, agentId string, requestData []byte) error {
	return CastPlayer(accountId, &RequestParams{agentId, requestData})
}

func HandleRPCCall(accountId string, requestData []byte) ([]byte, error) {
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
	return EncodeResponseData(reply.EncodeMethod, reqId, reply.Response)
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
		gen_server.Cast(self.PlayerId, &PersistData{})
	})
}

func (self *Player) HandleCast(msg interface{}) {
	switch params := msg.(type) {
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
	}
}

func (self *Player) HandleCall(msg interface{}) (interface{}, error) {
	switch params := msg.(type) {
	case *WrapParams:
		return self.handleWrap(params), nil
	case *RpcCallParams:
		return self.handleRPCCall(params)
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
	return self.sendDataToStream(agentId, 0, encode_method, msg)
}

type RequestParams struct {
	agentId string
	data []byte
}
func (self *Player) handleRequest(params *RequestParams) error {
	self.lastActive = time.Now().Unix()
	if !gosconf.IS_DEBUG {
		defer func() {
			if x := recover(); x != nil {
				logger.ERR("caught panic in player handleRequest(): ", x)
			}
		}()
	}

	reqId, handler, args, err := api.ParseRequestDataForHander(params.data)
	if err != nil {
		logger.ERR("ParseRequestDataForHander failed: ", err)
		return self.sendDataToStream(params.agentId, reqId, pt.PT_Fail, &pt.Fail{Fail: "error_route_not_found"})
	} else {
		encode_method, msg := self.processRequest(handler, args)
		return self.sendDataToStream(params.agentId, reqId, encode_method, msg)
	}
}

type RpcCallParams struct {
	Handler routes.Handler
	Params  interface{}
}
func (self *Player) handleRPCCall(params *RpcCallParams) (*RPCReply, error) {
	encode_method, response := params.Handler(self, params.Params)
	return &RPCReply{EncodeMethod: encode_method, Response: response}, nil
}

type RPCCastParams struct { data []byte }
func (self *Player) handleRPCCast(params *RPCCastParams) {
	_, handler, args, err := api.ParseRequestDataForHander(params.data)
	if err != nil {
		logger.ERR("handleRPCCast failed: ", err)
		return
	}
	self.processRequest(handler, args)
}

func EncodeResponseData(encode_method string, reqId int32, response interface{}) ([]byte, error) {
	writer, err := api.Encode(encode_method, response)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	return writer.GetSendData(reqId)
}

func (self *Player) processRequest(handler routes.Handler, params interface{}) (string, interface{}) {
	encode_method, response := handler(self, params)
	self.processed++
	logger.INFO("Processed: ", self.processed, " Response Data: ", response)
	return encode_method, response
}

type WrapParams struct { fun func(ctx *Player) interface{} }
func (self *Player) handleWrap(params *WrapParams) interface{} {
	self.lastActive = time.Now().Unix()
	return params.fun(self)
}

type AsyncWrapParams struct { fun func(ctx *Player) }
func (self *Player) handleAsyncWrap(params *AsyncWrapParams) {
	self.lastActive = time.Now().Unix()
	params.fun(self)
}

type PersistData struct {}
func (self *Player) handlePersistData() {
	err := player_data.Persist(self.PlayerId, self.Data)
	if err != nil {
		logger.ERR("PersistData failed: ", err)
	}
	self.startPersistTimer()
}

type BroadcastParams struct { msg *broadcast.BroadcastMsg }
func (self *Player) handleBroadcast(params *BroadcastParams) {
	if BroadcastHandler != nil {
		BroadcastHandler(self, params.msg)
	}
}

type ConnectedParams struct {
	agentId string
	agent interfaces.AgentBehavior
}
func (self *Player) handleConnected(params *ConnectedParams) {
	self.Agents[params.agentId] = params.agent
}

type DisConnectedParams struct { agentId string }
func (self *Player) handleDisconnected(params *DisConnectedParams) {
	delete(self.Agents, params.agentId)
}

/*
   IPC Methods
*/

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

func (self *Player) sendDataToStream(agentId string, reqId int32, encode_method string, msg interface{}) error {
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
	data, err := writer.GetSendData(reqId)
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
