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
	CALL        byte = 0
	CAST        byte = 1
	MANUAL_CALL byte = 2 // need manual response
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
	Category   byte
	ResultChan chan *Response
	Msg        interface{}
}

type GenServer struct {
	name        string
	callback    GenServerBehavior
	msgChannel  chan *Request
	signChannel chan *SignPacket
}

type GenServerBehavior interface {
	Init(args []interface{}) (err error)
	HandleCast(req *Request)
	HandleCall(req *Request) (interface{}, error)
	Terminate(reason string) (err error)
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{
			ResultChan: make(chan *Response),
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
		gen_server, err = New(module, args)
		if err != nil {
			return
		}
		setGenServer(server_name, gen_server)
	} else {
		fmt.Println(server_name, " is already exists!")
	}
	return
}

func New(module GenServerBehavior, args ...interface{}) (*GenServer, error) {
	msgChannel := make(chan *Request, 1024)
	signChannel := make(chan *SignPacket)

	gen_server := &GenServer{
		callback:    module,
		msgChannel:  msgChannel,
		signChannel: signChannel,
	}

	err := gen_server.callback.Init(args)
	if err != nil {
		logger.ERR("gen_server start failed: ", err)
		return nil, err
	}

	go loop(gen_server) // Enter infinity loop

	return gen_server, err
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
	return callByCategory(CALL, server_name, msg)
}

func ManualCall(server_name string, msg interface{}) (interface{}, error) {
	return callByCategory(MANUAL_CALL, server_name, msg)
}

func callByCategory(category byte, server_name string, msg interface{}) (interface{}, error) {
	if gen_server, exists := GetGenServer(server_name); exists {
		return gen_server.callByCategory(category, msg)
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

func (self *GenServer) Call(msg interface{}) (interface{}, error) {
	return self.callByCategory(CALL, msg)
}

func (self *GenServer) ManualCall(msg interface{}) (interface{}, error) {
	return self.callByCategory(MANUAL_CALL, msg)
}

func (self *GenServer) callByCategory(category byte, msg interface{}) (interface{}, error) {
	request := requestPool.Get().(*Request)
	defer func() {
		requestPool.Put(request)
	}()
	request.Category = category
	request.Msg = msg

	self.msgChannel <- request

	packet := <-request.ResultChan
	result := packet.result
	err := packet.err

	responsePool.Put(packet)

	return result, err
}

func (self *GenServer) Cast(msg interface{}) {
	request := requestPool.Get().(*Request)
	defer func() {
		requestPool.Put(request)
	}()
	request.Category = CAST
	request.Msg = msg
	self.msgChannel <- request
}

func (self *Request) Response(result interface{}, err error) {
	resp := responsePool.Get().(*Response)
	resp.result = result
	resp.err = err
	self.ResultChan <- resp
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
				switch req.Category {
				case CALL:
					result, err := gen_server.callback.HandleCall(req)
					req.Response(result, err)
					break
				case CAST:
					gen_server.callback.HandleCast(req)
					break
				case MANUAL_CALL:
					_, _ = gen_server.callback.HandleCall(req)
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
