package connection

import (
	"net"
	"time"
	"io"
	"encoding/binary"
	"goslib/packet"
	"goslib/logger"
	"api"
	"agent"
	"errors"
	"app/consts"
	"goslib/redisDB"
	"sync"
)

const TCP_TIMEOUT = 5 // seconds

type Session struct {
	AccountId string
	ServerId  string
	SceneId   string
	ConnectAppId string
	GameAppId string
	Token     string
	Connection *Connection
}

type Connection struct {
	authed bool
	conn net.Conn
	processed int64
	agent *agent.Agent
	session *Session
}

var sessionMap *sync.Map

func Start(_conn net.Conn) {
	sessionMap = new(sync.Map)
	instance := &Connection{
		authed: false,
		conn: _conn,
		processed: 0,
	}
	go instance.handleRequest()
}

func GetSession(accountId string) *Session {
	session, ok := sessionMap.Load(accountId)
	if !ok {
		return nil
	}
	return session.(*Session)
}

func (self *Connection)SendRawData(data []byte) {
	self.conn.Write(data)
}

func (self *Connection)handleRequest() {
	header := make([]byte, 2)

	for {
		data, err := self.receiveRequest(header)
		if err != nil {
			break
		}

		if self.authed {
			agent.ProxyToGame(self.session, data)
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
				self.session.Connection = self
				sessionMap.Store(self.session.AccountId, self.session)
				agent.DispatchToGameApp(self.session)
			}
		}
	}
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
	sessionMap, err := redisDB.Instance().HGetAll("session:" + params.AccountId).Result()
	if err != nil {
		return false, err
	}

	if sessionMap["token"] != params.Token {
		return false, errors.New("Session Token invalid!")
	}

	self.session = &Session{
		AccountId: params.AccountId,
		ServerId:  sessionMap["serverId"],
		SceneId:   sessionMap["sceneId"],
		Token:     params.Token,
	}

	return true,nil
}
