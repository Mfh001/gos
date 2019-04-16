package gen_server

import (
	"errors"
	"fmt"
	"goslib/logger"
	"sync"
)

const SIGN_STOP = 1

var ServerRegisterMap = sync.Map{}

const (
	CALL byte = 0
	CAST byte = 1
)

type Packet struct {
	method string
	args   []interface{}
}

type SignPacket struct {
	signal           int
	reason           string
	response_channel chan *Response
}

type Response struct {
	result interface{}
	err    error
}

type Request struct {
	category byte
	resultChan chan *Response
	msg interface{}
}

type GenServer struct {
	name        string
	callback    GenServerBehavior
	msgChannel  chan *Request
	signChannel chan *SignPacket
}

type GenServerBehavior interface {
	Init(args []interface{}) (err error)
	HandleCast(msg interface{})
	HandleCall(msg interface{}) (interface{}, error)
	Terminate(reason string) (err error)
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{
			resultChan: make(chan *Response),
		}
	},
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return &Response{}
	},
}

func setGenServer(name string, instance *GenServer) {
	ServerRegisterMap.Store(name, instance)
}

func GetGenServer(name string) (*GenServer, bool) {
	if v, ok := ServerRegisterMap.Load(name); ok {
		return v.(*GenServer), ok
	}
	return nil, false
}

func Exists(name string) bool {
	_, ok := ServerRegisterMap.Load(name)
	return ok
}

func delGenServer(name string) {
	ServerRegisterMap.Delete(name)
}

func Start(server_name string, module GenServerBehavior, args ...interface{}) (gen_server *GenServer, err error) {
	gen_server, exists := GetGenServer(server_name)
	if !exists {
		msgChannel := make(chan *Request, 1024)
		signChannel := make(chan *SignPacket)

		gen_server = &GenServer{
			name:        server_name,
			callback:    module,
			msgChannel:  msgChannel,
			signChannel: signChannel}

		if err = gen_server.callback.Init(args); err != nil {
			logger.ERR("gen_server start failed: ", err)
			return
		}

		go loop(gen_server) // Enter infinity loop

		setGenServer(server_name, gen_server)
	} else {
		fmt.Println(server_name, " is already exists!")
	}
	return
}

func Stop(server_name, reason string) error {
	if gen_server, exists := GetGenServer(server_name); exists {
		response_channel := make(chan *Response)
		gen_server.signChannel <- &SignPacket{
			signal:           SIGN_STOP,
			reason:           reason,
			response_channel: response_channel,
		}
		response := <-response_channel
		return response.err
	} else {
		logger.WARN(server_name, " not found!")
		return nil
	}
}

func Call(server_name string, msg interface{}) (interface{}, error) {
	if gen_server, exists := GetGenServer(server_name); exists {
		return gen_server.Call(msg)
	} else {
		errMsg := fmt.Sprintf("GenServer call failed: %s %s", server_name, " server not found!")
		logger.ERR(errMsg)
		return nil, errors.New(errMsg)
	}
}

func Cast(server_name string, msg interface{}) {
	if gen_server, exists := GetGenServer(server_name); exists {
		gen_server.Cast(msg)
	}
}

func (self *GenServer)Call(msg interface{}) (interface{}, error) {
	request := requestPool.Get().(*Request)
	defer func() {
		requestPool.Put(request)
	}()
	request.category = CALL
	request.msg = msg

	self.msgChannel <- request

	packet := <-request.resultChan
	result := packet.result
	err := packet.err

	responsePool.Put(packet)

	return result, err
}

func (self *GenServer)Cast(msg interface{}) {
	request := requestPool.Get().(*Request)
	defer func() {
		requestPool.Put(request)
	}()
	request.category = CAST
	request.msg = msg
	self.msgChannel <- request
}

func loop(gen_server *GenServer) {
	defer func() {
		terminate(gen_server)
	}()

	var sign_packet *SignPacket
	var req *Request
	var ok bool
	for {
		select {
		case req, ok = <-gen_server.msgChannel:
			if ok {
				switch req.category {
				case CALL:
					result, err := gen_server.callback.HandleCall(req.msg)
					resp := responsePool.Get().(*Response)
					resp.result = result
					resp.err = err
					req.resultChan <- resp
					break
				case CAST:
					gen_server.callback.HandleCast(req.msg)
					break
				}
			}
		case sign_packet, ok = <-gen_server.signChannel:
			if ok {
				switch sign_packet.signal {
				case SIGN_STOP:
					if err := gen_server.callback.Terminate(sign_packet.reason); err != nil {
						logger.ERR("GenServer stop failed: ", err)
						sign_packet.response_channel <- &Response{
							err: err,
						}
					} else {
						sign_packet.response_channel <- &Response{
							err: nil,
						}
						return
					}
				}
			}
		}
	}
}

func terminate(gen_server *GenServer) {
	close(gen_server.msgChannel)
	close(gen_server.signChannel)
	delGenServer(gen_server.name)
}
