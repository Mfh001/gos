package player

import (
	"api"
	"fmt"
	"goslib/gen_server"
	"goslib/logger"
	"gslib/routes"
	"goslib/memStore"
	"goslib/packet"
	"runtime"
	"time"
	"goslib/broadcast"
	"gslib"
	"goslib/sessionMgr"
	"gslib/sceneMgr"
)

type Player struct {
	PlayerId     string
	Session      *sessionMgr.Session
	processed    int
	Store        *memStore.MemStore
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
}

const EXPIRE_DURATION = 1800
var BroadcastHandler func(*Player, *broadcast.BroadcastMsg) = nil

func HandleRequest(accountId string, requestData []byte) {
	CastPlayer(accountId, "handleRequest", requestData)
}

func CallPlayer(accountId string, args ...interface{}) (interface{}, error) {
	if !gen_server.Exists(accountId) {
		StartPlayer(accountId)
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
	self.Store = memStore.New(self)
	self.lastActive = time.Now().Unix()
	self.startActiveCheck()
	self.startPersistTimer()

	session, err := sessionMgr.Find(self.PlayerId)
	if err != nil {
		logger.ERR("Player lookup session failed: ", self.PlayerId, " err: ", err)
	} else {
		self.Session = session
		sceneMgr.TryLoadScene(session.SceneId)
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
	} else if method_name == "handleWrap" {
		self.handleWrap(args[1].(func() interface{}))
	} else if method_name == "handleAsyncWrap" {
		self.handleAsyncWrap(args[0].(func()))
	} else if method_name == "PersistData" {
		self.Store.Persist([]string{"models"})
		self.startPersistTimer()
	} else if method_name == "removeConn" {
		//self.Conn = nil
	} else if method_name == "broadcast" {
		self.handleBroadcast(args[1].(*broadcast.BroadcastMsg))
	}
}

func (self *Player) HandleCall(args []interface{}) interface{} {
	method_name := args[0].(string)
	if method_name == "handleWrap" {
		return self.handleWrap(args[1].(func() interface{}))
	}
	return nil
}

func (self *Player) Terminate(reason string) (err error) {
	fmt.Println("callback Termiante!")
	self.activeTimer.Stop()
	self.persistTimer.Stop()
	self.Store.Persist([]string{self.PlayerId})
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
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := api.IdToName[protocol]
	handler, err := routes.Route(decode_method)
	self.processed++
	if err == nil {
		params := api.Decode(decode_method, reader)
		encode_method, response := handler(self, params)
		writer := api.Encode(encode_method, response)
		// INFO("Processed: ", self.processed, " Response Data: ", response_data)
		self.sendToClient(writer.GetSendData())
	} else {
		logger.ERR(err)
		writer := api.Encode("Fail", &api.Fail{Fail: "error_route_not_found"})
		self.sendToClient(writer.GetSendData())
	}
}

func (self *Player) sendToClient(data []byte) {
	connectAppId := GetConnectId(self.PlayerId)
	if connectAppId != "" {
		ProxyToConnect(connectAppId, self.PlayerId, data)
	}
}

func (self *Player) handleWrap(fun func() interface{}) interface{} {
	self.lastActive = time.Now().Unix()
	return fun()
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

func (self *Player) Wrap(targetPlayerId string, fun func() interface{}) (interface{}, error) {
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
