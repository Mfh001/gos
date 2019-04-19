package account

import (
	"fmt"
	"game_mgr"
	"github.com/go-redis/redis"
	"gosconf"
	"goslib/logger"
	"goslib/redisdb"
	"goslib/secure"
	"goslib/session_utils"
	"strconv"
)

const (
	ACCOUNT_GUEST = iota
	ACCOUNT_NORMAL
)

type Account struct {
	Uuid     string
	GroupId  string
	Category int
	Username string
	Password string
}

/*
 * Lookup Account
 */
func Lookup(username string) (*Account, error) {
	values, err := redisdb.Instance().HGetAll(username).Result()
	if err == redis.Nil || len(values) == 0 {
		return nil, nil
	}

	if err != nil {
		logger.INFO("Account Lookup Error: ", err)
		return nil, err
	}

	fmt.Println("values: ", values)

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
		logger.INFO("Create account failed: ", err)
		return nil, err
	}

	fmt.Println("Create: ", val)

	return &Account{
		Uuid:     username,
		GroupId:  groupId,
		Category: ACCOUNT_NORMAL,
		Username: username,
		Password: password,
	}, nil
}

func Delete(username string) {
	redisdb.Instance().Del(username)
}

/*
 * Check password is valid for this account
 */
func (self *Account) Auth(password string) bool {
	return self.Password == password
}

func (self *Account) ChangePassword(newPassword string) {
}

/*
 * RPC
 * request ConnectAppMgr dispatch connectApp for user connecting
 */
func (self *Account) Dispatch() (string, string, *session_utils.Session, error) {
	dispatchInfo, err := game_mgr.DispatchGame(gosconf.GS_ROLE_DEFAULT, self.Uuid, self.GroupId)
	if err != nil {
		logger.ERR("Dispatch account failed: ", err)
		return "", "", nil, err
	}

	session, err := session_utils.Find(self.Uuid)
	if err != nil {
		logger.ERR("Dispatch account find session failed: ", err)
		return "", "", nil, err
	}
	session.Token = secure.SessionToken()
	if err = session.Save(); err != nil {
		return "", "", nil, err
	}

	return dispatchInfo.AppHost, dispatchInfo.AppPort, session, nil
}
