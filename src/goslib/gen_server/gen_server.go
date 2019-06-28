/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package gen_server

import (
	"errors"
	"fmt"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/utils"
	"sync"
	"time"
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
	signal          int
	reason          string
	responseChannel chan *Response
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
			ResultChan: make(chan *Response, 1),
		}
	},
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return &Response{}
	},
}

func SetGenServer(name string, instance *GenServer) {
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

func DelGenServer(name string) {
	ServerRegisterMap.Delete(name)
}

func Start(serverName string, module GenServerBehavior, args ...interface{}) (*GenServer, error) {
	genServer, ok := GetGenServer(serverName)
	if !ok {
		genServer, err := New(module, args...)
		if err != nil {
			return nil, err
		}
		genServer.name = serverName
		SetGenServer(serverName, genServer)
		return genServer, nil
	} else {
		logger.WARN(serverName, " is already exists!")
		return genServer, nil
	}
}

func New(module GenServerBehavior, args ...interface{}) (*GenServer, error) {
	msgChannel := make(chan *Request, 1024)
	signChannel := make(chan *SignPacket)

	genServer := &GenServer{
		callback:    module,
		msgChannel:  msgChannel,
		signChannel: signChannel,
	}

	err := genServer.callback.Init(args)
	if err != nil {
		logger.ERR("gen_server start failed: ", err)
		return nil, err
	}

	go loop(genServer) // Enter infinity loop

	return genServer, err
}

func Stop(serverName, reason string) error {
	if genServer, exists := GetGenServer(serverName); exists {
		return genServer.Stop(reason)
	} else {
		logger.WARN(serverName, " not found!")
		return nil
	}
}

func Call(serverName string, msg interface{}) (interface{}, error) {
	return callByCategory(CALL, serverName, msg)
}

func ManualCall(serverName string, msg interface{}) (interface{}, error) {
	return callByCategory(MANUAL_CALL, serverName, msg)
}

func callByCategory(category byte, serverName string, msg interface{}) (interface{}, error) {
	if genServer, exists := GetGenServer(serverName); exists {
		return genServer.callByCategory(category, msg)
	} else {
		errMsg := fmt.Sprintf("GenServer call failed: %s %s", serverName, " server not found!")
		logger.ERR(errMsg)
		return nil, errors.New(errMsg)
	}
}

func Cast(serverName string, msg interface{}) {
	if genServer, exists := GetGenServer(serverName); exists {
		genServer.Cast(msg)
	}
}

func (self *GenServer) Call(msg interface{}) (interface{}, error) {
	return self.callByCategory(CALL, msg)
}

func (self *GenServer) ManualCall(msg interface{}) (interface{}, error) {
	return self.callByCategory(MANUAL_CALL, msg)
}

func (self *GenServer) callByCategory(category byte, msg interface{}) (interface{}, error) {
	request := getRequest()
	request.Category = category
	request.Msg = msg

	self.msgChannel <- request

	select {
	case packet := <-request.ResultChan:
		result := packet.result
		err := packet.err
		putResponse(packet)
		putRequest(request)
		return result, err
	case <-time.After(gosconf.GEN_SERVER_CALL_TIMEOUT):
		return nil, errors.New("gen_server call timeout: " + utils.StructToStr(msg))
	}
}

func (self *GenServer) Cast(msg interface{}) {
	request := getRequest()
	request.Category = CAST
	request.Msg = msg
	self.msgChannel <- request
}

func (self *GenServer) Stop(reason string) error {
	response_channel := make(chan *Response)
	self.signChannel <- &SignPacket{
		signal:          SIGN_STOP,
		reason:          reason,
		responseChannel: response_channel,
	}
	response := <-response_channel
	return response.err
}

func (self *Request) Response(result interface{}, err error) {
	resp := getResponse()
	resp.result = result
	resp.err = err
	self.ResultChan <- resp
}

func getRequest() *Request {
	return requestPool.Get().(*Request)
}

func putRequest(req *Request) {
	req.Category = 0
	req.Msg = nil
	requestPool.Put(req)
}

func getResponse() *Response {
	return responsePool.Get().(*Response)
}

func putResponse(rsp *Response) {
	rsp.result = nil
	rsp.err = nil
	responsePool.Put(rsp)
}

func loop(genServer *GenServer) {
	defer func() {
		logger.INFO("genServer terminate: ", genServer.name)
		terminate(genServer)
	}()

	var signPacket *SignPacket
	var req *Request
	var ok bool
	for {
		select {
		case req, ok = <-genServer.msgChannel:
			if ok {
				handleRequest(genServer, req)
			}
		case signPacket, ok = <-genServer.signChannel:
			if ok {
				if exit := handleCommand(genServer, signPacket); exit {
					return
				}
			}
		}
	}
}

func handleRequest(genServer *GenServer, req *Request) {
	defer utils.RecoverPanic(genServer.name)
	switch req.Category {
	case CALL:
		result, err := genServer.callback.HandleCall(req)
		req.Response(result, err)
		break
	case CAST:
		genServer.callback.HandleCast(req)
		putRequest(req)
		break
	case MANUAL_CALL:
		_, _ = genServer.callback.HandleCall(req)
		break
	}
}

func handleCommand(genServer *GenServer, signPacket *SignPacket) bool {
	defer utils.RecoverPanic(genServer.name)
	switch signPacket.signal {
	case SIGN_STOP:
		if err := genServer.callback.Terminate(signPacket.reason); err != nil {
			logger.ERR("GenServer stop failed: ", err)
			signPacket.responseChannel <- &Response{
				err: err,
			}
		} else {
			signPacket.responseChannel <- &Response{
				err: nil,
			}
			return true
		}
	}
	return false
}

func terminate(genServer *GenServer) {
	DelGenServer(genServer.name)
	close(genServer.msgChannel)
	close(genServer.signChannel)
}
