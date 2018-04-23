package gslib

import (
	"api"
	"fmt"
	"goslib/gen_server"
	. "goslib/logger"
	"gslib/routes"
	"goslib/memStore"
	"goslib/packet"
	"net"
	"runtime"
	"time"
	"goslib/broadcast"
)

type Player struct {
	PlayerId     string
	processed    int
	Conn         net.Conn
	Store        *memStore.MemStore
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
}

const EXPIRE_DURATION = 1800
var BroadcastHandler func(*Player, *broadcast.BroadcastMsg) = nil

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	name := args[0].(string)
	fmt.Println("server ", name, " started!")
	self.PlayerId = name
	self.Store = memStore.New(self)
	self.lastActive = time.Now().Unix()
	self.startActiveCheck()
	self.startPersistTimer()
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
		self.handleRequest(args[1].([]byte), args[2].(net.Conn))
	} else if method_name == "handleWrap" {
		self.handleWrap(args[1].(func() interface{}))
	} else if method_name == "handleAsyncWrap" {
		self.handleAsyncWrap(args[0].(func()))
	} else if method_name == "PersistData" {
		self.Store.Persist([]string{"models"})
		self.startPersistTimer()
	} else if method_name == "removeConn" {
		self.Conn = nil
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
	if self.Conn != nil {
		writer := api.Encode(encode_method, msg)
		writer.Send(self.Conn)
	}
}

func (self *Player) handleRequest(data []byte, conn net.Conn) {
	self.lastActive = time.Now().Unix()
	self.Conn = conn
	defer func() {
		if x := recover(); x != nil {
			ERR("caught panic in player handleRequest(): ", x)
		}
	}()
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := api.IdToName[protocol]
	handler, err := routes.Route(decode_method)
	if err == nil {
		params := api.Decode(decode_method, reader)
		encode_method, response := handler(self, params)
		writer := api.Encode(encode_method, response)

		self.processed++
		// INFO("Processed: ", self.processed, " Response Data: ", response_data)
		if self.Conn != nil {
			writer.Send(self.Conn)
		}
	} else {
		ERR(err)
		// TODO
		// response erro msg to user
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
		return gen_server.Call(targetPlayerId, "handleWrap", fun)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, fun func()) {
	if self.PlayerId == targetPlayerId {
		self.handleAsyncWrap(fun)
	} else {
		gen_server.Cast(targetPlayerId, "HandleAsyncWrap", fun)
	}
}

func (self *Player) JoinChannel(channel string) {
	gen_server.Cast(BROADCAST_SERVER_ID, "JoinChannel", self.PlayerId, channel)
}

func (self *Player) LeaveChannel(channel string) {
	gen_server.Cast(BROADCAST_SERVER_ID, "LeaveChannel", self.PlayerId, channel)
}

func (self *Player) PublishChannelMsg(channel, category string, data interface{}) {
	msg := &broadcast.BroadcastMsg{
		Category: category,
		Channel:  channel,
		SenderId: self.PlayerId,
		Data:     data,
	}
	gen_server.Cast(BROADCAST_SERVER_ID, "Publish", msg)
}
