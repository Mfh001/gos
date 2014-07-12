package main

import (
	"api"
	"fmt"
	"routes"
	"runtime"
	. "utils"
	"utils/packet"
)

type Player struct {
	processed int
	OutBuffer *Buffer
}

/*
   GenServer Callbacks
*/
func (self *Player) Init(name string) (err error) {
	fmt.Println("server ", name, " started!")
	return nil
}

func (self *Player) HandleCast(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "HandleRequest" {
		self.HandleRequest(args[1].([]byte), args[2].(*Buffer))
	}
}

func (self *Player) HandleCall(args []interface{}) {
	method_name := args[0].(string)
	if method_name == "HandleRPC" {
	}
}

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
