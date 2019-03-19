package player

import (
	"fmt"
	"gen/api/pt"
	"gen/proto"
	"github.com/kataras/iris/core/errors"
	"gosconf"
	"goslib/api"
	"goslib/broadcast"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/memstore"
	"goslib/packet"
	"goslib/routes"
	"goslib/session_utils"
	//"gslib"
	"goslib/scene_mgr"
	"runtime"
	"time"
)

type Player struct {
	PlayerId     string
	Store        *memstore.MemStore
	stream       proto.RouteConnectGame_AgentStreamServer
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

func PlayerConnected(accountId string, stream proto.RouteConnectGame_AgentStreamServer) {
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
	return EncodeResponseData(reply.EncodeMethod, reply.Response)
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
		self.stream = args[1].(proto.RouteConnectGame_AgentStreamServer)
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

func (self *Player) SendData(encode_method string, msg interface{}) error {
	writer, err := api.Encode(encode_method, msg)
	if err != nil {
		return err
	}
	return self.sendToClient(writer.GetSendData())
}

func (self *Player) handleRequest(data []byte) {
	self.lastActive = time.Now().Unix()
	if !gosconf.IS_DEBUG {
		defer func() {
			if x := recover(); x != nil {
				logger.ERR("caught panic in player handleRequest(): ", x)
			}
		}()
	}

	handler, params, err := ParseRequestData(data)
	if err != nil {
		logger.ERR(err)
		data, err := failMsgData("error_route_not_found")
		if err == nil {
			self.sendToClient(data)
		}
	} else {
		data, err := self.processRequest(handler, params)
		if err != nil {
			data, err := failMsgData("error_msg_encoding_failed")
			if err == nil {
				self.sendToClient(data)
			}
		} else {
			self.sendToClient(data)
		}
	}
}

func (self *Player) handleRPCCall(handler routes.Handler, params interface{}) (*RPCReply, error) {
	encode_method, response := handler(self, params)
	return &RPCReply{EncodeMethod: encode_method, Response: response}, nil
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
	decode_method := pt.IdToName[protocol]
	handler, err := routes.Route(decode_method)
	logger.INFO("handelRequest: ", decode_method)
	if err != nil {
		return nil, nil, err
	}
	params, err := api.Decode(decode_method, reader)
	return handler, params, err
}

func EncodeResponseData(encode_method string, response interface{}) ([]byte, error) {
	writer, err := api.Encode(encode_method, response)
	if err != nil {
		logger.ERR("EncodeResponseData failed: ", err)
		return nil, err
	}
	return writer.GetSendData(), nil
}

func (self *Player) processRequest(handler routes.Handler, params interface{}) ([]byte, error) {
	encode_method, response := handler(self, params)
	self.processed++
	logger.INFO("Processed: ", self.processed, " Response Data: ", response)
	return EncodeResponseData(encode_method, response)
}

func (self *Player) sendToClient(data []byte) error {
	if self.stream != nil {
		err := self.stream.Send(&proto.RouteMsg{
			Data: data,
		})
		if err != nil {
			logger.ERR("sendToClient failed: ", err)
			return err
		} else {
			return nil
		}
	} else {
		errMsg := "sendToClient failed, connectAppId is nil!"
		logger.WARN(errMsg)
		return errors.New(errMsg)
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

func (self *Player) JoinChannel(channel string) {
	broadcast.JoinChannel(self.PlayerId, channel)
}

func (self *Player) LeaveChannel(channel string) {
	broadcast.LeaveChannel(self.PlayerId, channel)
}

func (self *Player) PublishChannelMsg(channel, category string, data interface{}) {
	broadcast.PublishChannelMsg(self.PlayerId, channel, category, data)
}

func failMsgData(errorMsg string) ([]byte, error) {
	writer, err := api.Encode("Fail", &pt.Fail{Fail: errorMsg})
	if err != nil {
		logger.ERR("Encode msg failed: ", err)
		return nil, err
	}
	return writer.GetSendData(), nil
}
