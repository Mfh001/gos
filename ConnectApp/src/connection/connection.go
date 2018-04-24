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
	"app/consts"
	"sync"
	"goslib/sessionMgr"
)

const TCP_TIMEOUT = 5 // seconds

type Connection struct {
	authed bool
	conn net.Conn
	processed int64
	agent *Agent
	session *sessionMgr.Session
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

func GetSession(accountId string) *sessionMgr.Session {
	session, ok := sessionMap.Load(accountId)
	if !ok {
		return nil
	}
	return session.(*sessionMgr.Session)
}

func SendRawData(accountId string, data []byte) {
	connection, ok := connectionMap.Load(accountId)
	if ok {
		connection.(*Connection).conn.Write(data)
	}
}

func (self *Connection)handleRequest() {
	header := make([]byte, 2)

	for {
		data, err := self.receiveRequest(header)
		if err != nil {
			break
		}

		if self.authed {
			ProxyToGame(self.session, data)
		} else {
			success, err := self.authConn(data)
			if err != nil {
				logger.ERR("AuthConn Error: ", err)
				break
			}
			self.authed = success
			if success {
				if err != nil {
					logger.ERR("GetRoleInfo error: ", err)
					break
				}
				sessionMap.Store(self.session.AccountId, self.session)
				connectionMap.Store(self.session.AccountId, self)
				DispatchToGameApp(self.session)
				SetRegisterAccountToGameApp(self.session, true)
			}
		}
	}

	SetRegisterAccountToGameApp(self.session, false)
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
	data := make([]byte, size)
	n, err = io.ReadFull(self.conn, data)
	if err != nil {
		logger.ERR("error receiving msg, bytes: ", n, "reason: ", err)
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

	params := api.Decode(decode_method, reader).(*consts.SessionAuthParams)

	// Validate Token from AuthApp
	success, err := self.validateSession(params)
	if err != nil {
		return false, err
	}

	// Send auth response to client
	writer := api.Encode("SessionAuthResponse", &consts.SessionAuthResponse{Success: success})

	self.processed++
	// INFO("Processed: ", self.processed, " Response Data: ", response_data)
	if self.conn != nil {
		writer.Send(self.conn)
	}

	return success, nil
}

func (self *Connection)validateSession(params *consts.SessionAuthParams) (bool, error) {
	// Get session from redis
	session, err := sessionMgr.Find(params.AccountId)
	if err != nil {
		return false, err
	}

	if session.Token != params.Token {
		return false, errors.New("Session Token invalid!")
	}

	self.session = session

	return true,nil
}
