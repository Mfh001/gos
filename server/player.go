package main

import (
	"api"
	"fmt"
	"routes"
	"runtime"
	. "utils"
)

type Player struct {
	processed int
}

// gen_server callbacks
func (self *Player) Init(name string) (err error) {
	fmt.Println("server ", name, " started!")
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

func (self *Player) HandleRequest(data []byte, out *Buffer) {
	protocol := 1
	// remain_data := []byte("hello")
	controller, action, err := routes.Route(protocol)
	if err == nil {
		decode_method := "DecodeEquipsUnloadParams"
		args := CallWithArgs(new(api.Decoder), decode_method, data)
		response_args := Call(controller, action, args)
		encode_method := response_args[0].String()
		response_data := Call(new(api.Encoder), encode_method, response_args[1:])
		self.processed++
		if err = out.Send([]byte("hello")); err != nil {
			ERR("cannot send to client", err)
		}
		INFO("Processed: ", self.processed, " Response Data: ", response_data)
	} else {
		ERR("error routes not found")
	}
}
