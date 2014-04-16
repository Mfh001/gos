package routes

import (
	"app/controllers"
	"errors"
)

type Routes struct {
}

type Router struct {
	controller interface{}
	method     string
}

var routes = make(map[int]Router)

func InitRoutes() {
	controller := new(controllers.EquipsController)
	routes[1] = Router{controller, "Load"}
}

func Route(protocol int) (interface{}, string, error) {
	router, ok := routes[protocol]
	if ok {
		return router.controller, router.method, nil
	} else {
		return nil, "", errors.New("Router not found!")
	}
}
