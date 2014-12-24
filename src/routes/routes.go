package routes

import (
	"errors"
	// "reflect"
)

type Handler func(ctx interface{}, params interface{}) interface{}

var routes = map[uint16]Handler{}

func Add(protocol uint16, handler Handler) {
	routes[protocol] = handler
}

func Route(protocol uint16) (Handler, error) {
	handler, ok := routes[protocol]
	if ok {
		return handler, nil
	} else {
		return handler, errors.New("Router not found!")
	}
}
