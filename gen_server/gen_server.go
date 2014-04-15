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
	cast_channel chan []reflect.Value
	call_channel chan []reflect.Value
	sign_channel chan SignPacket
}

var SIGN_STOP int = 1

type GenServerBehavior interface {
	// Init(args []reflect.Value) (err error)
	// HandleCall(method string, args []reflect.Value) []reflect.Value
	// HandleCast(method string, args []reflect.Value)
	Terminate(reason string) (err error)
}

var NamingMap = make(map[string]GenServer)

func Start(server_name string, module GenServerBehavior, args ...interface{}) (gen_server GenServer) {
	if _, exists := NamingMap[server_name]; !exists {
		cast_channel := make(chan []reflect.Value, 1024)
		call_channel := make(chan []reflect.Value)
		sign_channel := make(chan SignPacket)
		gen_server = GenServer{
			name:         server_name,
			callback:     module,
			cast_channel: cast_channel,
			call_channel: call_channel,
			sign_channel: sign_channel}
		// gen_server.callback.Init(to_reflect_values(args)) // Init callback struct
		utils.Call(gen_server.callback, "Init", to_reflect_values(args))
		go loop(gen_server) // Enter infinity loop
		NamingMap[server_name] = gen_server
	} else {
		fmt.Println(server_name, " is already exists!")
	}
	return NamingMap[server_name]
}

func Stop(server_name, reason string) {
	if gen_server, exists := NamingMap[server_name]; exists {
		gen_server.sign_channel <- SignPacket{SIGN_STOP, reason}
	} else {
		fmt.Println(server_name, " not found!")
	}
}

func Call(server_name string, args ...interface{}) (result []reflect.Value, err error) {
	if gen_server, exists := NamingMap[server_name]; exists {
		gen_server.call_channel <- to_reflect_values(args)
		result = <-gen_server.call_channel
	} else {
		fmt.Println(server_name, " not found!")
		err = errors.New("Server not found!")
	}
	return
}

func Cast(server_name string, args ...interface{}) {
	if gen_server, exists := NamingMap[server_name]; exists {
		gen_server.cast_channel <- to_reflect_values(args)
	} else {
		fmt.Println(server_name, " not found!")
	}
}

func loop(gen_server GenServer) {
	for {
		select {
		case args, ok := <-gen_server.cast_channel:
			if ok {
				// fmt.Println("handle_cast: ", args)
				method := args[0].String()
				utils.Call(gen_server.callback, method, args[1:])
				// gen_server.callback.HandleCast(method, args[1:])
			}
		case args, ok := <-gen_server.call_channel:
			if ok {
				// fmt.Println("handle_call: ", args)
				method := args[0].String()
				// result := gen_server.callback.HandleCall(method, args[1:])
				result := utils.Call(gen_server.callback, method, args[1:])
				gen_server.call_channel <- result
			}
		case sign_packet, ok := <-gen_server.sign_channel:
			if ok {
				// fmt.Println("handle_sign: ", sign_packet)
				switch sign_packet.signal {
				case SIGN_STOP:
					gen_server.callback.Terminate(sign_packet.reason)
					close(gen_server.cast_channel)
					close(gen_server.call_channel)
					close(gen_server.sign_channel)
					delete(NamingMap, gen_server.name)
					break
				}
			}
		}
	}
}

func to_reflect_values(args []interface{}) []reflect.Value {
	in := make([]reflect.Value, len(args))
	for k, arg := range args {
		in[k] = reflect.ValueOf(arg)
	}
	return in
}

/*
   callback module usage:
   type Player struct {
       id string
   }

   func (p *Player) HandleCall(args interface{}, from chan int) (reply interface{}) {
   }

   var player = Player{id: "uuid-001"}
   var gen_server GenServerBehavior = player
   gen_server.HandleCall(...)
   gen_server.HandleCast(...)
*/
