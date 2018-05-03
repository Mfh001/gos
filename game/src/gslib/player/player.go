package player

import (
	"api"
	"fmt"
	"github.com/kataras/iris/core/errors"
	pb "gos_rpc_proto"
	"goslib/broadcast"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/memstore"
	"goslib/packet"
	"goslib/session_utils"
	"gslib"
	"gslib/routes"
	"gslib/scene_mgr"
	"runtime"
	"time"
)

type Player struct {
	PlayerId     string
	Store        *memstore.MemStore
	Session      *session_utils.Session
	stream       pb.RouteConnectGame_AgentStreamServer
	processed    int
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
}

type RPCReply struct {
	encode_method string
	response      interface{}
}

const EXPIRE_DURATION = 1800

var BroadcastHandler func(*Player, *broadcast.BroadcastMsg) = nil
var CurrentGameAppId string

func PlayerConnected(accountId string, stream pb.RouteConnectGame_AgentStreamServer) {
	CastPlayer(accountId, "connected", stream)
}

func PlayerDisconnected(accountId string) {
	CastPlayer(accountId, "disconnected")
}

func HandleRequest(accountId string, requestData []byte) {
	CastPlayer(accountId, "handleRequest", requestData)
}

func HandleRPCCall(accountId string, requestData []byte) ([]byte, error) {
	handler, params, err := ParseRequestData(requestData)
	if err != nil {
		return nil, err
	}
	result, err := CallPlayer(accountId, "handleRPCCall", handler, params)
	if err != nil {
		logger.ERR("HandleRPCCall failed: ", err)
		return nil, err
	}
	reply := result.(*RPCReply)
	return EncodeResponseData(reply.encode_method, reply.response), nil
}

func HandleRPCCast(accountId string, requestData []byte) {
	CastPlayer(accountId, "handleRPCCast", requestData)
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

func CastPlayer(accountId string, args ...interface{}) {
	if !gen_server.Exists(accountId) {
		StartPlayer(accountId)
	}
	gen_server.Cast(accountId, args...)
}

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	name := args[0].(string)
	fmt.Println("Player: ", name, " started!")
	self.PlayerId = name
	self.Store = memstore.New(name, self)
	self.lastActive = time.Now().Unix()
	self.startActiveCheck()
	self.startPersistTimer()

	session, err := session_utils.Find(self.PlayerId)
	if err != nil {
		logger.ERR("Player lookup session failed: ", self.PlayerId, " err: ", err)
	} else {
		self.Session = session
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
		self.handleRequest(args[1].([]byte))
	} else if method_name == "handleRPCCast" {
		self.handleRPCCast(args[1].([]byte))
	} else if method_name == "handleWrap" {
		self.handleWrap(args[1].(func(player *Player) interface{}))
	} else if method_name == "handleAsyncWrap" {
		self.handleAsyncWrap(args[0].(func()))
	} else if method_name == "PersistData" {
		self.Store.Persist([]string{"models"})
		self.startPersistTimer()
	} else if method_name == "removeConn" {
		//self.Conn = nil
	} else if method_name == "broadcast" {
		self.handleBroadcast(args[1].(*broadcast.BroadcastMsg))
	} else if method_name == "connected" {
		self.stream = args[1].(pb.RouteConnectGame_AgentStreamServer)
	} else if method_name == "disconnected" {
		self.stream = nil
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
	self.Store.Persist([]string{"models"})
	if ok := memstore.EnsurePersisted(self.PlayerId); !ok {
		return errors.New("Persist player data failed!")
	}
	return nil
}

func (self *Player) startActiveCheck() {
	if (self.lastActive + EXPIRE_DURATION) < time.Now().Unix() {
		gen_server.Stop(self.PlayerId, "Shutdown inActive player!")
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

func (self *Player) SendData(encode_method string, msg interface{}) {
	writer := api.Encode(encode_method, msg)
	self.sendToClient(writer.GetSendData())
}

func (self *Player) handleRequest(data []byte) {
	self.lastActive = time.Now().Unix()
	defer func() {
		if x := recover(); x != nil {
			logger.ERR("caught panic in player handleRequest(): ", x)
		}
	}()

	handler, params, err := ParseRequestData(data)
	if err != nil {
		logger.ERR(err)
		self.sendToClient(failMsgData("error_route_not_found"))
	} else {
		data := self.processRequest(handler, params)
		self.sendToClient(data)
	}
}

func (self *Player) handleRPCCall(handler routes.Handler, params interface{}) (*RPCReply, error) {
	encode_method, response := handler(self, params)
	return &RPCReply{encode_method: encode_method, response: response}, nil
}

func (self *Player) handleRPCCast(data []byte) {
	handler, params, err := ParseRequestData(data)
	if err != nil {
		logger.ERR("handleRPCCast failed: ", err)
		return
	}
	self.processRequest(handler, params)
}

func ParseRequestData(data []byte) (routes.Handler, interface{}, error) {
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := api.IdToName[protocol]
	handler, err := routes.Route(decode_method)
	logger.INFO("handelRequest: ", decode_method)
	if err != nil {
		return nil, nil, err
	}
	params := api.Decode(decode_method, reader)
	return handler, params, nil
}

func EncodeResponseData(encode_method string, response interface{}) []byte {
	writer := api.Encode(encode_method, response)
	return writer.GetSendData()
}

func parseResponseData(data []byte) interface{} {
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := api.IdToName[protocol]
	logger.INFO("handelResponse: ", decode_method)
	params := api.Decode(decode_method, reader)
	return params
}

func (self *Player) processRequest(handler routes.Handler, params interface{}) []byte {
	encode_method, response := handler(self, params)
	self.processed++
	logger.INFO("Processed: ", self.processed, " Response Data: ", response)
	return EncodeResponseData(encode_method, response)
}

func (self *Player) sendToClient(data []byte) {
	if self.stream != nil {
		err := self.stream.Send(&pb.RouteMsg{
			Data: data,
		})
		if err != nil {
			logger.ERR("sendToClient failed: ", err)
		}
	} else {
		logger.WARN("sendToClient failed, connectAppId is nil!")
	}
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
		CastPlayer(targetPlayerId, "HandleAsyncWrap", fun)
	}
}

func (self *Player) RequestPlayer(targetPlayerId string, encode_method string, params interface{}) (interface{}, error) {
	if gen_server.Exists(targetPlayerId) {
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	session, err := session_utils.Find(targetPlayerId)
	if err != nil {
		return nil, err
	}
	if self.Session.GameAppId == session.GameAppId {
		return internalRequestPlayer(targetPlayerId, encode_method, params)
	}
	return crossRequestPlayer(session, encode_method, params)
}

func (self *Player) JoinChannel(channel string) {
	gen_server.Cast(gslib.BROADCAST_SERVER_ID, "JoinChannel", self.PlayerId, channel)
}

func (self *Player) LeaveChannel(channel string) {
	gen_server.Cast(gslib.BROADCAST_SERVER_ID, "LeaveChannel", self.PlayerId, channel)
}

func (self *Player) PublishChannelMsg(channel, category string, data interface{}) {
	msg := &broadcast.BroadcastMsg{
		Category: category,
		Channel:  channel,
		SenderId: self.PlayerId,
		Data:     data,
	}
	gen_server.Cast(gslib.BROADCAST_SERVER_ID, "Publish", msg)
}

func failMsgData(errorMsg string) []byte {
	writer := api.Encode("Fail", &api.Fail{Fail: errorMsg})
	return writer.GetSendData()
}
