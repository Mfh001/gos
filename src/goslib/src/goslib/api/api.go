package api

import (
	"errors"
	"gen/api/jsonapi"
	"gen/api/pbapi"
	"gen/api/pt"
	"gen/api/rawapi"
	"gosconf"
	"goslib/logger"
	"goslib/packet"
	"goslib/routes"
	"strconv"
)

func Encode(encode_method string, v interface{}) (*packet.Packet, error) {
	switch gosconf.AGENT_ENCODING {
	case gosconf.PROTOCOL_ENCODING_PB:
		buffer, err := pbapi.Encode(encode_method, v)
		return buffer, err
	case gosconf.PROTOCOL_ENCODING_RAW:
		buffer := rawapi.Encode(encode_method, v)
		return buffer, nil
	case gosconf.PROTOCOL_ENCODING_JSON:
		buffer, err := jsonapi.Encode(encode_method, v)
		return buffer, err
	default:
		return nil, errors.New("unsupport AGENT_ENCODING: " + strconv.Itoa(gosconf.AGENT_ENCODING))
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

func ParseRequestData(data []byte) (string, interface{}, error) {
	reader := packet.Reader(data)
	protocol, err := reader.ReadUint16()
	if err != nil {
		return "", nil, err
	}
	decode_method := pt.IdToName[protocol]
	params, err := Decode(decode_method, reader)
	return decode_method, params, err
}

func ParseRequestDataForHander(data []byte) (routes.Handler, interface{}, error) {
	decode_method, params, err := ParseRequestData(data)
	if err != nil {
		return nil, nil, err
	}
	logger.INFO("handelRequest: ", decode_method)
	handler, err := routes.Route(decode_method)
	if err != nil {
		return nil, nil, err
	}
	return handler, params, err
}

func ParseProtolType(data []byte) (int, error) {
	reader := packet.Reader(data)
	protocol, err := reader.ReadUint16()
	if err != nil {
		logger.ERR("parseProtocolType failed: ", err)
		return 0, err
	}
	protocolType := pt.IdToType[protocol]
	return protocolType, nil
}
