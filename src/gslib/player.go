package gslib

import (
	"api"
	"fmt"
	"gen_server"
	"routes"
	"runtime"
	. "utils"
	"utils/packet"
)

type Player struct {
	playerId  string
	processed int
	OutBuffer *Buffer
}

// gen_server callbacks
func (self *Player) Init(name string) (err error) {
	fmt.Println("server ", name, " started!")
	self.playerId = name
	return nil
}

// gen_server callbacks
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
		packet.Pack(protocol, response_struct, writer)
		response_data := writer.Data()
		self.processed++
		// INFO("Processed: ", self.processed, " Response Data: ", response_data)
		if err = out.Send(response_data); err != nil {
			ERR("cannot send to client", err)
		}
	} else {
		ERR(err)
	}
}

func (self *Player) HandleWrap(fun func()) {
	fun()
}

func (self *Player) Wrap(targetPlayerId string, fun func()) {
	if self.playerId == targetPlayerId {
		self.HandleWrap(fun)
	} else {
		gen_server.Call(targetPlayerId, "HandleWrap", fun)
	}
}

func (self *Player) AsyncWrap(targetPlayerId string, fun func()) {
	if self.playerId == targetPlayerId {
		self.HandleWrap(fun)
	} else {
		gen_server.Cast(targetPlayerId, "HandleWrap", fun)
	}
}
