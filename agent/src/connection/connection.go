package connection

import (
	"api"
	"encoding/binary"
	"errors"
	pb "gos_rpc_proto"
	"gosconf"
	"goslib/game_utils"
	"goslib/logger"
	"goslib/packet"
	"goslib/session_utils"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	id        int
	authed    bool
	conn      net.Conn
	processed int64
	session   *session_utils.Session
	stream    pb.RouteConnectGame_AgentStreamClient
}

var sessionMap = &sync.Map{}
var connectionMap = &sync.Map{}
var onlinePlayers int32
var connId = 0

func OnlinePlayers() int32 {
	return onlinePlayers
}

func Handle(_conn net.Conn) {
	instance := &Connection{
		id:        connId,
		authed:    false,
		conn:      _conn,
		processed: 0,
	}
	connectionMap.Store(instance.id, instance)
	connId++
	atomic.AddInt32(&onlinePlayers, 1)
	go instance.handleRequest()
}

func (self *Connection) handleRequest() {
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
				logger.ERR("DispatchToGameApp failed: ", err)
				break
			}
		}
	}
	atomic.AddInt32(&onlinePlayers, -1)
	if self.stream != nil {
		err := self.stream.CloseSend()
		if err != nil {
			logger.ERR("Connection close stream failed: ", err)
		}
	}
}

func (self *Connection) setupProxy() error {
	sessionMap.Store(self.session.AccountId, self.session)
	var game *game_utils.Game
	var err error
	if self.session.GameAppId != "" {
		game, err = game_utils.Find(self.session.GameAppId)
		if err != nil {
			return err
		}
	} else {
		game, err = ChooseGameServer(self.session)
		if err != nil {
			return err
		}
	}
	self.stream, err = ConnectGameServer(game.Uuid, self.session.AccountId, self.conn)
	if err != nil {
		logger.ERR("StartConnToGameStream failed: ", err)
		return err
	}
	return nil
}

func (self *Connection) proxyRequest(data []byte) error {
	if self.stream == nil {
		return errors.New("Stream not exits!")
	}
	err := self.stream.Send(&pb.RouteMsg{
		Data: data,
	})
	return err
}

// Block And Receiving "request data"
func (self *Connection) receiveRequest(header []byte) ([]byte, error) {
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

func (self *Connection) authConn(data []byte) (bool, error) {
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

func (self *Connection) validateSession(params *api.SessionAuthParams) (bool, error) {
	session, err := session_utils.Find(params.AccountId)
	if err != nil {
		return false, err
	}

	if session.Token != params.Token {
		return false, errors.New("Session token invalid!")
	}

	self.session = session

	return true, nil
}
