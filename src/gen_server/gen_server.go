package gen_server

import (
	"errors"
	"fmt"
	"reflect"
	"utils"
)

type Packet struct {
	method string
	args   []interface{}
}

type SignPacket struct {
	signal int
	reason string
}

type GenServer struct {
	name         string
	callback     GenServerBehavior
	// cast_channel chan []reflect.Value
	cast_channel chan []interface{}
	call_channel chan []reflect.Value
	sign_channel chan SignPacket
}

var SIGN_STOP int = 1

type GenServerBehavior interface {
	// Init(args ...interface{}) (err error)
    HandleCast(args []interface{})
	Terminate(reason string) (err error)
}

func Start(server_name string, module GenServerBehavior, args ...interface{}) (gen_server GenServer) {
	gen_server, exists := GetGenServer(server_name)
	if !exists {
		// cast_channel := make(chan []reflect.Value, 1024)
		cast_channel := make(chan []interface{}, 1024)
		call_channel := make(chan []reflect.Value)
		sign_channel := make(chan SignPacket)

		gen_server = GenServer{
			name:         server_name,
			callback:     module,
			cast_channel: cast_channel,
			call_channel: call_channel,
			sign_channel: sign_channel}

		utils.Call(gen_server.callback, "Init", utils.ToReflectValues(args))

		go loop(gen_server) // Enter infinity loop

		SetGenServer(server_name, gen_server)
	} else {
		fmt.Println(server_name, " is already exists!")
	}
	return gen_server
}

func Stop(server_name, reason string) {
	if gen_server, exists := GetGenServer(server_name); exists {
		gen_server.sign_channel <- SignPacket{SIGN_STOP, reason}
	} else {
		fmt.Println(server_name, " not found!")
	}
}

func Call(server_name string, args ...interface{}) (result []reflect.Value, err error) {
	if gen_server, exists := GetGenServer(server_name); exists {
		response_channel := make(chan []reflect.Value)
		defer func() {
			close(response_channel)
		}()
		args = append(args, response_channel)
		gen_server.call_channel <- utils.ToReflectValues(args)
		result = <-response_channel
	} else {
		fmt.Println(server_name, " not found!")
		err = errors.New("Server not found!")
	}
	return
}

func (self *GenServer) Call(args ...interface{}) (result []reflect.Value, err error) {
	response_channel := make(chan []reflect.Value)
	defer func() {
		close(response_channel)
	}()
	args = append(args, response_channel)
	self.call_channel <- utils.ToReflectValues(args)
	result = <-response_channel
	return
}

func Cast(server_name string, args ...interface{}) {
	if gen_server, exists := GetGenServer(server_name); exists {
		// gen_server.cast_channel <- utils.ToReflectValues(args)
		gen_server.cast_channel <- args
	} else {
		fmt.Println(server_name, " not found!")
	}
}

func (self *GenServer) Cast(args ...interface{}) {
    self.cast_channel <- args
}

func loop(gen_server GenServer) {
	defer func() {
		terminate(gen_server)
	}()

	for {
		select {
		case args, ok := <-gen_server.cast_channel:
			if ok {
				// utils.INFO("handle_cast: ", args)
				// method := args[0].String()
				// utils.Call(gen_server.callback, method, args[1:])
				// gen_server.callback.HandleCast(method, args[1:])
				gen_server.callback.HandleCast(args)
			}
		case args, ok := <-gen_server.call_channel:
			if ok {
				// utils.INFO("handle_call: ", args)
				method := args[0].String()
				size := len(args)
				response_channel := args[size-1]
				result := utils.Call(gen_server.callback, method, args[1:size-1])
				response_channel.Send(reflect.ValueOf(result))
			}
		case sign_packet, ok := <-gen_server.sign_channel:
			if ok {
				// utils.INFO("handle_sign: ", sign_packet)
				switch sign_packet.signal {
				case SIGN_STOP:
					gen_server.callback.Terminate(sign_packet.reason)
					return
				}
			}
		}
	}
}

func terminate(gen_server GenServer) {
	close(gen_server.cast_channel)
	close(gen_server.call_channel)
	close(gen_server.sign_channel)
	DelGenServer(gen_server.name)
}
