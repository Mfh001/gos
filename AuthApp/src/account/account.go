package account

import (
	"time"
	"context"
	"log"
	pb "connectAppProto"
	"redisDB"
	"fmt"
	"strconv"
)

var ConnectRpcClient pb.DispatcherClient

const (
	ACCOUNT_GUEST = iota
	ACCOUNT_NORMAL
)

type Account struct {
	Uuid string
	Category int
	Username string
	Password string
}

/*
 * Lookup Account
 */
func Lookup(accountId string) *Account {
	values, err := redisDB.Instance().HMGet(accountId, "uuid", "category", "username", "password").Result()
	if err != nil {
		log.Fatalf("Account Lookup Error: %v", err)
		return nil
	}

	fmt.Println("values: ", values)

	if values[0] == nil {
		return nil
	}

	category, err := strconv.Atoi(values[1].(string))

	return &Account{
		values[0].(string),
		category,
		values[2].(string),
		values[3].(string),
	}
}

/*
 * Register Account
 */
func Create(username string, password string) *Account {
	params := make(map[string]interface{})
	params["uuid"] = username
	params["category"] = ACCOUNT_NORMAL
	params["username"] = username
	params["password"] = password

	val, err := redisDB.Instance().HMSet(username, params).Result()

	if err != nil {
		log.Fatalf("Create account failed: %v", err)
	}

	fmt.Println("Create: ", val)

	return &Account{
		Uuid:username,
		Category:ACCOUNT_NORMAL,
		Username:username,
		Password:password,
	}
}

func Delete(accountId string) {
	redisDB.Instance().Del(accountId)
}

/*
 * Check password is valid for this account
 */
func (self *Account)Auth(password string) bool {
	return true
}

func (self *Account)ChangePassword(newPassword string) {
}

/*
 * RPC
 * request ConnectAppMgr dispatch connectApp for user connecting
 */
func (self *Account)Dispatch() (connectAppHost string, connectAppPort string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := ConnectRpcClient.DispatchPlayer(ctx, &pb.DispatchRequest{AccountId:self.Uuid})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return
	}

	log.Printf("Greeting: %s:%s", reply.GetConnectAppHost(), reply.GetConnectAppPort())

	connectAppHost = reply.GetConnectAppHost()
	connectAppPort = reply.GetConnectAppPort()

	return
}
