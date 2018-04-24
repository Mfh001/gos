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
		log.Fatalf("Account Lookup Error: %v", err)
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
	params["uuid"] = username
	params["groupId"] = ""
	params["category"] = ACCOUNT_NORMAL
	params["username"] = username
	params["password"] = password

	fmt.Println("uuid: ", params["uuid"])
	val, err := redisDB.Instance().HMSet(username, params).Result()

	if err != nil {
		log.Fatalf("Create account failed: %v", err)
		return nil, err
	}

	fmt.Println("Create: ", val)

	return &Account{
		Uuid:username,
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
func (self *Account)Dispatch() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := ConnectRpcClient.DispatchPlayer(ctx, &pb.DispatchRequest{
		AccountId:self.Uuid,
		GroupId:self.GroupId,
		//GroupId:strconv.Itoa(rand.Intn(20)),
		})
	if err != nil {
		logger.ERR("could not greet: ", err)
		return "", "", err
	}

	logger.DEBUG("Greeting: %s:%s", reply.GetConnectAppHost(), reply.GetConnectAppPort())

	session, err := sessionMgr.Find(self.Uuid)
	if err != nil {
		return "", "", err
	}

	if session == nil {
		_, err := sessionMgr.Create(map[string]string{
			"accountId": self.Uuid,
			"serverId": self.GroupId,
			"sceneId": "",
			"connectAppId": reply.GetConnectAppId(),
			"gameAppId": "",
			"token": secure.SessionToken(),
		})
		if err != nil {
			return "", "", nil
		}
	} else {
		session.ConnectAppId = reply.GetConnectAppId()
		session.Save()
	}

	return reply.GetConnectAppHost(), reply.GetConnectAppPort(), nil
}
