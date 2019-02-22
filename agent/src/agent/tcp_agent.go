package main

import (
	"connection"
	"encoding/binary"
	"github.com/gorilla/websocket"
	"gosconf"
	"goslib/logger"
	"io"
	"net"
	"time"
)

type TCPAgent struct {
	mt      int
	tcpConn *websocket.Conn
	connIns *connection.Connection
}

func StartTCPAgent() {
	// Listen incomming tcp connections
	tcpConf := gosconf.TCP_SERVER_CONNECT_APP
	l, err := net.Listen("tcp", tcpConf.Address)
	if err != nil {
		logger.ERR("Connection listen failed: ", err)
	}
	logger.INFO("ConnectApp started!")
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			logger.ERR("Connection accept failed: ", err)
		}

		go tcpHandler(conn)
	}
}

func tcpHandler(conn net.Conn) {
	agent := &TCPAgent{}
	connIns := connection.New(agent)

	defer func() {
		connIns.Cleanup()
	}()

	agent.connIns = connIns
	agent.tcpConn = conn
	header := make([]byte, gosconf.TCP_SERVER_CONNECT_APP.Packet)
	for {
		data, err := agent.receiveRequest(header)
		if err != nil {
			break
		}
		if err = agent.onMessage(data); err != nil {
			break
		}
	}
}

// Block And Receiving "request data"
func (self *TCPAgent) receiveRequest(header []byte) ([]byte, error) {
	self.tcpConn.SetReadDeadline(time.Now().Add(gosconf.TCP_READ_TIMEOUT))
	_, err := io.ReadFull(self.tcpConn, header)
	if err != nil {
		logger.ERR("Receive data head failed: ", err)
		return nil, err
	}

	size := binary.BigEndian.Uint16(header)
	data := make([]byte, size)
	_, err = io.ReadFull(self.tcpConn, data)
	if err != nil {
		logger.ERR("Receive data body failed: ", err)
		return nil, err
	}
	return data, nil
}

func (self *TCPAgent) onMessage(data []byte) error {
	logger.INFO("ws received: ", data)
	err := self.connIns.OnMessage(data)
	return err
}

func (self *TCPAgent) SendMessage(data []byte) error {
	err := self.tcpConn.WriteMessage(self.mt, data)
	if err != nil {
		logger.ERR("write: ", err)
		return err
	}
	return nil
}
