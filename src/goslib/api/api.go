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

package api

import (
	"errors"
	"github.com/mafei198/gos/goslib/gen/api/jsonapi"
	"github.com/mafei198/gos/goslib/gen/api/pbapi"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	"github.com/mafei198/gos/goslib/gen/api/rawapi"
	"github.com/mafei198/gos/goslib/gosconf"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/packet"
	"github.com/mafei198/gos/goslib/routes"
	"github.com/mafei198/gos/goslib/utils"
	"strconv"
)

func Encode(v interface{}) (*packet.Packet, error) {
	switch gosconf.AGENT_ENCODING {
	case gosconf.PROTOCOL_ENCODING_PB:
		buffer, err := pbapi.Encode(v)
		return buffer, err
	case gosconf.PROTOCOL_ENCODING_RAW:
		buffer, err := rawapi.Encode(v)
		return buffer, err
	case gosconf.PROTOCOL_ENCODING_JSON:
		buffer, err := jsonapi.Encode(v)
		return buffer, err
	default:
		return nil, errors.New("unsupport AGENT_ENCODING: " + strconv.Itoa(gosconf.AGENT_ENCODING))
	}
}

func EncodeMethod(v interface{}) (method string) {
	switch gosconf.AGENT_ENCODING {
	case gosconf.PROTOCOL_ENCODING_PB:
		method, _ = pbapi.EncodeMethod(v)
		return
	case gosconf.PROTOCOL_ENCODING_RAW:
		method, _ = rawapi.EncodeMethod(v)
		return
	case gosconf.PROTOCOL_ENCODING_JSON:
		method, _ = jsonapi.EncodeMethod(v)
		return
	default:
		return ""
	}
}

func Decode(decode_method string, buffer *packet.Packet) (interface{}, error) {
	switch gosconf.AGENT_ENCODING {
	case gosconf.PROTOCOL_ENCODING_PB:
		params, err := pbapi.Decode(decode_method, buffer)
		return params, err
	case gosconf.PROTOCOL_ENCODING_RAW:
		params, err := rawapi.Decode(decode_method, buffer)
		return params, err
	case gosconf.PROTOCOL_ENCODING_JSON:
		params, err := jsonapi.Decode(decode_method, buffer)
		return params, err
	default:
		return nil, errors.New("unsupport AGENT_ENCODING: " + strconv.Itoa(gosconf.AGENT_ENCODING))
	}
}

func ParseRequestData(data []byte) (int32, string, interface{}, error) {
	reader := packet.Reader(data)
	reqId, err := reader.ReadInt32()
	if err != nil {
		return 0, "", nil, err
	}
	protocol, err := reader.ReadUint16()
	if err != nil {
		return 0, "", nil, err
	}
	decode_method := pt.IdToName[protocol]
	params, err := Decode(decode_method, reader)
	return reqId, decode_method, params, err
}

func ParseRequestDataForHander(data []byte) (int32, routes.Handler, interface{}, error) {
	reqId, decode_method, params, err := ParseRequestData(data)
	if err != nil {
		return 0, nil, nil, err
	}
	logger.INFO("handelRequest reqId: ", reqId, " path: ", decode_method, " params: ", utils.StructToStr(params))
	handler, err := routes.Route(decode_method)
	if err != nil {
		return 0, nil, nil, err
	}
	return reqId, handler, params, err
}

func ParseProtolType(data []byte) (int, error) {
	reader := packet.Reader(data)
	_, err := reader.ReadInt32()
	protocol, err := reader.ReadUint16()
	if err != nil {
		logger.ERR("parseProtocolType failed: ", err)
		return 0, err
	}
	protocolType := pt.IdToType[protocol]
	return protocolType, nil
}
