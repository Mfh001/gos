/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package account_utils

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"strconv"
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
func Create(username, groupId string, password string, category int) (*Account, error) {
	params := make(map[string]interface{})
	params["uuid"] = username
	params["groupId"] = groupId
	params["category"] = category
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
		Category: category,
		Username: username,
		Password: password,
	}, nil
}

func CreateService(username, groupId string) (*Account, error) {
	return Create(username, groupId, "", 0)
}

func Delete(username string) {
	redisdb.Instance().Del(username)
}
