package connection

import (
	"net"
	"time"
	"io"
	"encoding/binary"
	"goslib/packet"
	"goslib/logger"
	"api"
	"errors"
	"sync"
	"goslib/session_mgr"
	pb "gos_rpc_proto"
	"gosconf"
)

type Connection struct {
	id int
	authed bool
	conn net.Conn
	processed int64
	session *session_mgr.Session
	stream pb.RouteConnectGame_AgentStreamClient
}

var sessionMap *sync.Map
var connectionMap *sync.Map
var connId = 0

func Handle(_conn net.Conn) {
	sessionMap = new(sync.Map)
	instance := &Connection{
		id: connId,
		authed: false,
		conn: _conn,
		processed: 0,
	}
	connectionMap.Store(instance.id, instance)
	connId++
	go instance.handleRequest()
}

func (self *Connection)handleRequest() {
	defer func() {
		connectionMap.Delete(self.id)
		self.stream = nil
		self.conn = nil
		self.session = nil
		self.authed = false
	}()

	header := make([]byte, 2)

	for {
		data, err := self.receiveRequest(header)
		if err != nil {
			break
		}

		if self.authed {
			if err = self.proxyRequest(data); err != nil {
				logger.ERR("SendMsg to GameApp failed: ", self.session.AccountId, " err: ", err)
				break
			}
		} else {
			if self.authed, err = self.authConn(data); err != nil {
				logger.ERR("AuthConn failed: ", err)
				break
			}
			if !self.authed {
				logger.ERR("AuthConn failed! ")
				break
			}
			if err = self.setupProxy(); err != nil {
				break
			}
		}
	}
}

func (self *Connection)setupProxy() error {
	sessionMap.Store(self.session.AccountId, self.session)
	err := ChooseGameServer(self.session)
	if err != nil {
		logger.ERR("DispatchToGameApp failed: ", err)
		return err
	}
	self.stream, err = ConnectGameServer(self.session.GameAppId, self.session.AccountId, self.conn)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return err
	}
	return nil
}

func (self *Connection)proxyRequest(data []byte) error {
	if self.stream == nil {
		return errors.New("Stream not exits!")
	}
	err := self.stream.Send(&pb.RouteMsg{
		Data: data,
	})
	return err
}

// Block And Receiving "request data"
func (self *Connection)receiveRequest(header []byte) ([]byte, error) {
	self.conn.SetReadDeadline(time.Now().Add(gosconf.TCP_READ_TIMEOUT))
	_, err := io.ReadFull(self.conn, header)
	if err != nil {
		logger.ERR("Receive data head failed: ", err)
		return nil, err
	}

	size := binary.BigEndian.Uint16(header)
	data := make([]byte, size)
	_, err = io.ReadFull(self.conn, data)
	if err != nil {
		logger.ERR("Receive data body failed: ", err)
		return nil, err
	}
	return data, nil
}

func (self *Connection)authConn(data []byte) (bool, error){
	reader := packet.Reader(data)
	protocol := reader.ReadUint16()
	decode_method := api.IdToName[protocol]

	if decode_method != "SessionAuthParams" {
		return false, errors.New("Request UnAuthed connection: " + decode_method)
	}

	params := api.Decode(decode_method, reader).(*api.SessionAuthParams)

	// Validate Token from AuthApp
	success, err := self.validateSession(params)
	if err != nil {
		return false, err
	}

	// Send auth response to client
	writer := api.Encode("SessionAuthResponse", &api.SessionAuthResponse{Success: success})

	self.processed++
	// INFO("Processed: ", self.processed, " Response Data: ", response_data)
	if self.conn != nil {
		writer.Send(self.conn)
	}

	return success, nil
}

func (self *Connection)validateSession(params *api.SessionAuthParams) (bool, error) {
	session, err := session_mgr.Find(params.AccountId)
	if err != nil {
		return false, err
	}

	if session.Token != params.Token {
		return false, errors.New("Session token invalid!")
	}

	self.session = session

	return true,nil
}
