package gslib

import (
	"api"
	"fmt"
	"gslib/gen_server"
	"gslib/routes"
	"gslib/store"
	. "gslib/utils"
	"gslib/utils/packet"
	"net"
	"runtime"
	"time"
)

type Player struct {
	PlayerId     string
	processed    int
	Conn         net.Conn
	Store        *store.Ets
	activeTimer  *time.Timer
	persistTimer *time.Timer
	lastActive   int64
}

type WrapHandler func() interface{}
type AsyncWrapHandler func()

const (
	EXPIRE_DURATION = 1800
)

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	name := args[0].(string)
	fmt.Println("server ", name, " started!")
	self.PlayerId = name
	self.Store = store.New()
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
	if method_name == "HandleRequest" {
		self.HandleRequest(args[1].([]byte), args[2].(net.Conn))
	} else if method_name == "HandleWrap" {
		self.HandleWrap(args[1].(WrapHandler))
	} else if method_name == "PersistData" {
		self.Store.Persist([]string{self.PlayerId})
		self.startPersistTimer()
	} else if method_name == "removeConn" {
		self.Conn = nil
	}
}

func (self *Player) HandleCall(args []interface{}) interface{} {
	method_name := args[0].(string)
	if method_name == "HandleWrap" {
		return self.HandleWrap(args[1].(WrapHandler))
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

func (self *Player) HandleRequest(data []byte, conn net.Conn) {
	self.lastActive = time.Now().Unix()
	self.Conn = conn
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("caught panic in player HandleRequest(): ", x)
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
	}
}

func (self *Player) HandleWrap(fun WrapHandler) interface{} {
	self.lastActive = time.Now().Unix()
	return fun()
}

func (self *Player) HandleAsyncWrap(fun AsyncWrapHandler) {
	self.lastActive = time.Now().Unix()
	fun()
}

func (self *Player) Wrap(targetPlayerId string, fun WrapHandler) (interface{}, error) {
	if self.PlayerId == targetPlayerId {
		return self.HandleWrap(fun), nil
	} else {
		return gen_server.Call(targetPlayerId, "HandleWrap", fun)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, fun AsyncWrapHandler) {
	if self.PlayerId == targetPlayerId {
		self.HandleAsyncWrap(fun)
	} else {
		gen_server.Cast(targetPlayerId, "HandleAsyncWrap", fun)
	}
}
