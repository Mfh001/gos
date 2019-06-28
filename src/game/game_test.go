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
package main

import (
	"github.com/mafei198/gos/game/app/custom_register/callbacks"
	"github.com/mafei198/gos/game/app/gen_register"
	"github.com/mafei198/gos/goslib/gen/api/pt"
	. "github.com/onsi/ginkgo"
	"github.com/mafei198/gos/goslib/api"
	"github.com/mafei198/gos/goslib/game_server"
	"github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player"
	"github.com/mafei198/gos/goslib/redisdb"
	"github.com/mafei198/gos/goslib/session_utils"
	"net"
	"time"
)

var _ = Describe("Game", func() {

	redisdb.StartClient()
	go game_server.Start("GS", func() {
		gen_register.RegisterRoutes()
		callbacks.RegisterSceneLoad()
	}, nil, nil)

	accountId := "fakeAccountId"
	token := "fake_token"
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:4000")
		if err != nil {
			logger.ERR("connect socket failed: ", err)
			time.Sleep(2 * time.Second)
			continue
		}
		session_utils.Create(&session_utils.Session{
			AccountId: accountId,
			GameAppId: player.CurrentGameAppId,
			Token:     token,
		})
		writer, err := api.Encode(&pt.SessionAuthParams{
			AccountId: accountId,
			Token:     token,
		})
		data, _ := writer.GetSendData(0)
		conn.Write(data)

		writer, err = api.Encode(&pt.EquipLoadParams{
			PlayerID: accountId,
			EquipId:  "fakeeid",
			HeroId:   "fakehero",
		})
		data, _ = writer.GetSendData(0)
		conn.Write(data)

		time.Sleep(1000 * time.Second)
		break
	}
})
