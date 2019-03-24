package agent

import (
	"github.com/gorilla/websocket"
	"gosconf"
	"goslib/game_server/connection"
	"goslib/logger"
	"net"
	"net/http"
	"sync/atomic"
)

type WSAgent struct {
	mt      int
	wsConn  *websocket.Conn
	connIns *connection.Connection
}

func StartWSAgent() {
	http.HandleFunc("/", wsHandler)

	tcpConf := gosconf.TCP_SERVER_GAME

	l, err := net.Listen("tcp", tcpConf.Address)
	if err != nil {
		logger.ERR("Connection listen failed: ", err)
		panic(err)
	}
	logger.INFO("WSAgent lis: ", tcpConf.Address)
	go http.Serve(l, nil)
}

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger.INFO("WSAgent accepted new conn")
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

	atomic.AddInt32(&OnlinePlayers, 1)
	for {
		mt, message, err := c.ReadMessage()
		agent.mt = mt
		if err != nil {
			logger.ERR("read: ", err)
			break
		}
		if err = agent.OnMessage(message); err != nil {
			break
		}
	}
	atomic.AddInt32(&OnlinePlayers, -1)
	agent.OnDisconnected()
}

func (self *WSAgent) OnMessage(data []byte) error {
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

func (self *WSAgent) OnDisconnected() {
	self.connIns.OnDisconnected()
}
