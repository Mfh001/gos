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
package account

import (
	"github.com/mafei198/gos/goslib/account_utils"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/scene_utils"
	"github.com/mafei198/gos/goslib/secure"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/world/game_mgr"
)

const (
	ACCOUNT_GUEST = iota
	ACCOUNT_NORMAL
)

/*
 * Register Account
 */
func Create(username string, password string, category int) (*account_utils.Account, error) {
	groupId, err := ChooseServer()
	if err != nil {
		logger.ERR("ChooseServer failed: ", err)
		return nil, err
	}
	return account_utils.Create(username, groupId, password, category)
}

/*
 * 分配服务器
 */
func ChooseServer() (string, error) {
	scenes, err := scene_utils.LoadAll()
	if err != nil {
		return "", err
	}
	var currentScene *scene_utils.Scene
	for _, scene := range scenes {
		if currentScene == nil {
			currentScene = scene
		} else if scene.Registered < currentScene.Registered {
			currentScene = scene
		}
	}
	return currentScene.Uuid, nil
}

func Delete(username string) {
	redisdb.Instance().Del(username)
}

/*
 * Check password is valid for this account
 */
func Auth(acc *account_utils.Account, password string) bool {
	return acc.Password == password
}

/*
 * RPC
 * request ConnectAppMgr dispatch connectApp for user connecting
 */
func Dispatch(acc *account_utils.Account) (string, string, *session_utils.Session, error) {
	dispatchInfo, err := game_mgr.DispatchGame(acc.Uuid, acc.GroupId)
	if err != nil {
		logger.ERR("Dispatch account failed: ", err)
		return "", "", nil, err
	}

	session, err := session_utils.Find(acc.Uuid)
	if err != nil {
		logger.ERR("Dispatch account find session failed: ", err)
		return "", "", nil, err
	}
	session.Token = secure.SessionToken()
	if err = session.Save(); err != nil {
		return "", "", nil, err
	}
	session_utils.Active(session.AccountId)

	return dispatchInfo.AppHost, dispatchInfo.AppPort, session, nil
}
