package main

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

func (self *Player) SystemInfo(from string, time int) int {
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
	controller, action, err := routes.Route(protocol)
	controller.Context = self
	if err == nil {
		decode_method := "DecodeEquipsUnloadParams"
		args := CallWithArgs(new(api.Decoder), decode_method, reader)
		response_args := Call(controller, action, args)
		encode_method := response_args[0].String()
		response := Call(new(api.Encoder), encode_method, response_args[1:])
		self.processed++
		response_data := response[0].Elem().FieldByName("data").Bytes()
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
