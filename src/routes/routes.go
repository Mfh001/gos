package routes

import (
	"app/controllers"
	"errors"
)

type Router struct {
	controller interface{}
	method     string
}

type Handler interface{}

var routes = map[uint16]Handler{
	1: func(params ...interface{}) { controllers.EquipsController.Load(params) },
}

func Route(protocol uint16) (interface{}, string, error) {
	router, ok := routes[protocol]
	if ok {
		return router.controller, router.method, nil
	} else {
		return nil, "", errors.New("Router not found!")
	}
}
