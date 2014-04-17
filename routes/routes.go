package routes

import (
	"app/controllers"
	"errors"
)

type Router struct {
	controller interface{}
	method     string
}

var routes = map[uint16]Router{
	1: Router{new(controllers.EquipsController), "Load"},
}

func Route(protocol uint16) (interface{}, string, error) {
	router, ok := routes[protocol]
	if ok {
		return router.controller, router.method, nil
	} else {
		return nil, "", errors.New("Router not found!")
	}
}
