package main

import (
	"connection"
	"flag"
	"github.com/gorilla/websocket"
	"goslib/logger"
	"net/http"
)

type WSAgent struct {
	mt      int
	wsConn  *websocket.Conn
	connIns *connection.Connection
}

func StartWSAgent() {
	http.HandleFunc("/", wsHandler)

	addr := flag.String("addr", "localhost:8080", "http service address")
	http.ListenAndServe(*addr, nil)
}

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ERR("upgrade:", err)
		return
	}
	defer c.Close()

	agent := &WSAgent{}
	connIns := connection.New(agent)
	agent.connIns = connIns
	agent.wsConn = c

	defer func() {
		connIns.Cleanup()
	}()

	for {
		mt, message, err := c.ReadMessage()
		agent.mt = mt
		if err != nil {
			logger.ERR("read: ", err)
			break
		}
		if err = agent.onMessage(message); err != nil {
			break
		}
	}
}

func (self *WSAgent) onMessage(data []byte) error {
	logger.INFO("ws received: ", data)
	err := self.connIns.OnMessage(data)
	return err
}

func (self *WSAgent) SendMessage(data []byte) error {
	err := self.wsConn.WriteMessage(self.mt, data)
	if err != nil {
		logger.ERR("write: ", err)
		return err
	}
	return nil
}
