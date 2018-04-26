package account

import (
	"time"
	"context"
	"log"
	pb "gos_rpc_proto"
	"goslib/redisdb"
	"fmt"
	"strconv"
	"goslib/logger"
	"goslib/session_mgr"
	"goslib/secure"
)

var ConnectRpcClient pb.DispatcherClient

const (
	ACCOUNT_GUEST = iota
	ACCOUNT_NORMAL
)

type Account struct {
	Uuid string
	GroupId string
	Category int
	Username string
	Password string
}

/*
 * Lookup Account
 */
func Lookup(username string) (*Account, error) {
	values, err := redisdb.Instance().HGetAll(username).Result()
	if err != nil {
		log.Println("Account Lookup Error: %v", err)
		return nil, err
	}

	fmt.Println("values: ", values)

	if values["uuid"] == "" {
		return nil, nil
	}

	category, err := strconv.Atoi(values["category"])

	return &Account{
		values["uuid"],
		values["groupId"],
		category,
		values["username"],
		values["password"],
	}, nil
}

/*
 * Register Account
 */
func Create(username string, password string) (*Account, error) {
	params := make(map[string]interface{})
	groupId := "server001"
	params["uuid"] = username
	params["groupId"] = groupId
	params["category"] = ACCOUNT_NORMAL
	params["username"] = username
	params["password"] = password

	fmt.Println("uuid: ", params["uuid"])
	val, err := redisdb.Instance().HMSet(username, params).Result()

	if err != nil {
		log.Println("Create account failed: %v", err)
		return nil, err
	}

	fmt.Println("Create: ", val)

	return &Account{
		Uuid:username,
		GroupId:groupId,
		Category:ACCOUNT_NORMAL,
		Username:username,
		Password:password,
	}, nil
}

func Delete(username string) {
	redisdb.Instance().Del(username)
}

/*
 * Check password is valid for this account
 */
func (self *Account)Auth(password string) bool {
	return self.Password == password
}

func (self *Account)ChangePassword(newPassword string) {
}

/*
 * RPC
 * request ConnectAppMgr dispatch connectApp for user connecting
 */
func (self *Account)Dispatch() (string, string, *session_mgr.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := ConnectRpcClient.DispatchPlayer(ctx, &pb.DispatchRequest{
		AccountId:self.Uuid,
		GroupId:self.GroupId,
		})
	if err != nil {
		logger.ERR("Dispatch account failed: ", err)
		return "", "", nil, err
	}

	session, err := self.updateSession(reply)

	logger.INFO(session.Uuid, session.ServerId, session.Token)

	return reply.GetConnectAppHost(), reply.GetConnectAppPort(), session, nil
}

func (self *Account)updateSession(reply *pb.DispatchReply) (*session_mgr.Session, error) {
	var session *session_mgr.Session
	var err error
	session, err = session_mgr.Find(self.Uuid)
	if err != nil {
		return nil, err
	}
	if session == nil {
		session, err = self.createSession()
		if err != nil {
			return nil, err
		}
	}
	session.ConnectAppId = reply.GetConnectAppId()
	err = session.Save()
	if err != nil {
		return nil, err
	}
	return session, err
}

func (self *Account)createSession() (*session_mgr.Session, error) {
	return session_mgr.Create(map[string]string{
		"accountId": self.Uuid,
		"serverId": self.GroupId,
		"sceneId": "",
		"connectAppId": "",
		"gameAppId": "",
		"token": secure.SessionToken(),
	})
}
