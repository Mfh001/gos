package gslib

import (
	"api"
	"fmt"
	"gslib/gen_server"
	"gslib/routes"
	. "gslib/utils"
	"gslib/utils/packet"
	"runtime"
)

type Player struct {
	playerId  string
	processed int
	OutBuffer *Buffer
}

type WrapHandler func() interface{}
type AsyncWrapHandler func()

/*
   GenServer Callbacks
*/
func (self *Player) Init(args []interface{}) (err error) {
	name := args[0].(string)
	fmt.Println("server ", name, " started!")
	self.playerId = name
	return nil
}

func (self *Player) HandleCast(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "HandleRequest" {
		self.HandleRequest(args[1].([]byte), args[2].(*Buffer))
	} else if method_name == "HandleWrap" {
		self.HandleWrap(args[1].(WrapHandler))
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
	return nil
}

/*
   IPC Methods
*/

func (self *Player) SystemInfo() int {
	return runtime.NumCPU()
}

func (self *Player) SendData(struct_name string, struct_instance interface{}) {
	var protocol int16 = 1
	writer := packet.Writer()
	data := packet.Pack(protocol, struct_instance, writer)
	self.OutBuffer.Send(data)
}

func (self *Player) HandleRequest(data []byte, out *Buffer) {
	self.OutBuffer = out
	reader := packet.Reader(data)
	protocol, _ := reader.ReadU16()
	handler, err := routes.Route(protocol)
	if err == nil {
		decode_method := "DecodeEquipsUnloadParams"
		params := api.Decode(decode_method, reader)
		response_struct := handler(self, params)

		writer := packet.Writer()
		var protocol int16 = 2
		response_data := packet.Pack(protocol, response_struct, writer)
		self.processed++
		// INFO("Processed: ", self.processed, " Response Data: ", response_data)
		if err = out.Send(response_data); err != nil {
			ERR("cannot send to client", err)
		}
	} else {
		ERR(err)
	}
}

func (self *Player) HandleWrap(fun WrapHandler) interface{} {
	return fun()
}

func (self *Player) HandleAsyncWrap(fun AsyncWrapHandler) {
	fun()
}

func (self *Player) Wrap(targetPlayerId string, fun WrapHandler) (interface{}, error) {
	if self.playerId == targetPlayerId {
		return self.HandleWrap(fun), nil
	} else {
		return gen_server.Call(targetPlayerId, "HandleWrap", fun)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, fun AsyncWrapHandler) {
	if self.playerId == targetPlayerId {
		self.HandleAsyncWrap(fun)
	} else {
		gen_server.Cast(targetPlayerId, "HandleAsyncWrap", fun)
	}
}
