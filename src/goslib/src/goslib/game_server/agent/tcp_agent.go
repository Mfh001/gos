package agent

import (
	"encoding/binary"
	"gosconf"
	"goslib/game_server/connection"
	"goslib/logger"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type TCPAgent struct {
	mt      int
	tcpConn net.Conn
	connIns *connection.Connection
}

func StartTCPAgent() {
	tcpConf := gosconf.TCP_SERVER_GAME
	l, err := net.Listen("tcp", tcpConf.Address)
	if err != nil {
		logger.ERR("Connection listen failed: ", err)
		panic(err)
	}
	logger.INFO("TcpAgent lis: ", tcpConf.Address)

	go acceptor(l)
}

func acceptor(l net.Listener) {
	defer l.Close()

	logger.INFO("Game TCPAgent started!")

	for {
		conn, err := l.Accept()
		logger.INFO("TCPAgent accepted new conn")
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
	header := make([]byte, gosconf.TCP_SERVER_GAME.Packet)

	atomic.AddInt32(&OnlinePlayers, 1)
	for {
		data, err := agent.receiveRequest(header)
		if err != nil {
			break
		}
		if err = agent.OnMessage(data); err != nil {
			break
		}
	}
	atomic.AddInt32(&OnlinePlayers, -1)

	agent.OnDisconnected()
}

// Block And Receiving "request data"
func (self *TCPAgent) receiveRequest(header []byte) ([]byte, error) {
	err := self.tcpConn.SetReadDeadline(time.Now().Add(gosconf.TCP_READ_TIMEOUT))
	if err != nil {
		logger.ERR("Receive data timeout: ", err)
		return nil, err
	}

	_, err = io.ReadFull(self.tcpConn, header)
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

func (self *TCPAgent) OnMessage(data []byte) error {
	logger.INFO("tcp received: ", data)
	err := self.connIns.OnMessage(data)
	return err
}

func (self *TCPAgent) SendMessage(data []byte) error {
	logger.INFO("Begin SendMessage")
	n, err := self.tcpConn.Write(data)
	if err != nil {
		logger.ERR("write: ", err)
		return err
	}
	logger.INFO("Finish SendMessage: ", n)
	return nil
}

func (self *TCPAgent) OnDisconnected() {
	self.connIns.OnDisconnected()
}
