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
	"goslib/sessionMgr"
	pb "gosRpcProto"
)

const TCP_TIMEOUT = 5 // seconds

type Connection struct {
	authed bool
	conn net.Conn
	processed int64
	session *sessionMgr.Session
	stream pb.RouteConnectGame_AgentStreamClient
}

var sessionMap *sync.Map
var connectionMap *sync.Map

func Start(_conn net.Conn) {
	sessionMap = new(sync.Map)
	connectionMap = new(sync.Map)
	instance := &Connection{
		authed: false,
		conn: _conn,
		processed: 0,
	}
	go instance.handleRequest()
}

func (self *Connection)handleRequest() {
	header := make([]byte, 2)

	for {
		data, err := self.receiveRequest(header)
		if err != nil {
			break
		}

		if self.authed {
			if self.stream != nil {
				err := self.stream.Send(&pb.RouteMsg{
					Data: data,
				})
				if err != nil {
					logger.ERR("SendMsg to GameApp failed: ", self.session.AccountId, " err: ", err)
					break
				}
			} else {
				logger.ERR("Stream not exist!")
			}
		} else {
			success, err := self.authConn(data)
			if err != nil {
				logger.ERR("AuthConn Error: ", err)
				break
			}
			self.authed = success
			if success {
				sessionMap.Store(self.session.AccountId, self.session)
				connectionMap.Store(self.session.AccountId, self)
				err = DispatchToGameApp(self.session)
				if err != nil {
					logger.ERR("DispatchToGameApp failed: ", err)
					break
				}
				self.stream, err = StartConnToGameStream(self.session.GameAppId, self.session.AccountId, self.conn)
				if err != nil {
					logger.ERR("StartConnToGameStream failed: ", err)
					break
				}
			}
		}
	}

	// save close connection
	self.stream.CloseSend()
	self.conn.Close()
	self.session = nil
	self.authed = false
}

// Block And Receiving "request data"
func (self *Connection)receiveRequest(header []byte) ([]byte, error) {
	self.conn.SetReadDeadline(time.Now().Add(TCP_TIMEOUT * time.Second))
	n, err := io.ReadFull(self.conn, header)
	if err != nil {
		logger.ERR("Connection Closed: ", err)
		return nil, err
	}

	size := binary.BigEndian.Uint16(header)
	logger.INFO("begin received: ", size)
	data := make([]byte, size)
	n, err = io.ReadFull(self.conn, data)
	if err != nil {
		logger.ERR("error receiving msg, bytes: ", n, "reason: ", err)
		return nil, err
	}
	logger.INFO("received: ", size)
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
	// Get session from redis
	session, err := sessionMgr.Find(params.AccountId)
	if err != nil {
		return false, err
	}

	logger.INFO("params.AccountId: ", params.AccountId, " params.Token: ", params.Token)
	logger.INFO("sessionId: ", session.Uuid, " sessionServerId: ", session.ServerId)
	if session.Token != params.Token {
		return false, errors.New("Session Token invalid!")
	}

	self.session = session

	return true,nil
}
