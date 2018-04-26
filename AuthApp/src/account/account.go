package account

import (
	"time"
	"context"
	"log"
	pb "gosRpcProto"
	"goslib/redisDB"
	"fmt"
	"strconv"
	"goslib/logger"
	"goslib/sessionMgr"
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
	values, err := redisDB.Instance().HGetAll(username).Result()
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
	val, err := redisDB.Instance().HMSet(username, params).Result()

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
	redisDB.Instance().Del(username)
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
func (self *Account)Dispatch() (string, string, *sessionMgr.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := ConnectRpcClient.DispatchPlayer(ctx, &pb.DispatchRequest{
		AccountId:self.Uuid,
		GroupId:self.GroupId,
		//GroupId:strconv.Itoa(rand.Intn(20)),
		})
	if err != nil {
		logger.ERR("could not greet: ", err)
		return "", "", nil, err
	}

	var session *sessionMgr.Session
	session, err = sessionMgr.Find(self.Uuid)
	if err != nil {
		return "", "", nil, err
	}

	if session == nil {
		logger.INFO("session not exists!")
		session, err = sessionMgr.Create(map[string]string{
			"accountId": self.Uuid,
			"serverId": self.GroupId,
			"sceneId": "",
			"connectAppId": reply.GetConnectAppId(),
			"gameAppId": "",
			"token": secure.SessionToken(),
		})
		if err != nil {
			return "", "", nil, nil
		}
	} else {
		logger.INFO("session exists!")
		session.ConnectAppId = reply.GetConnectAppId()
		session.Save()
	}

	logger.INFO(session.Uuid, session.ServerId, session.Token)

	return reply.GetConnectAppHost(), reply.GetConnectAppPort(), session, nil
}
