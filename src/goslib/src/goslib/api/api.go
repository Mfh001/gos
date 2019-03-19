package api

import (
	"errors"
	"gen/api/jsonapi"
	"gen/api/pbapi"
	"gen/api/rawapi"
	"gosconf"
	"goslib/packet"
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
		params := rawapi.Decode(decode_method, buffer)
		return params, nil
	case gosconf.PROTOCOL_ENCODING_JSON:
		params, err := jsonapi.Decode(decode_method, buffer)
		return params, err
	default:
		return nil, errors.New("unsupport AGENT_ENCODING: " + strconv.Itoa(gosconf.AGENT_ENCODING))
	}
}
