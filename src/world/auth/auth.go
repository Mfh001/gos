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
package auth

/*
验证服务器

功能：
	1.账户注册
	2.登录验证
	3.连接服分配
*/

import (
	"context"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/mafei198/gos/goslib/account_utils"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gosconf"
	gl "github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/session_utils"
	"github.com/mafei198/gos/world/auth/account"
	"net/http"
	"time"
)

var app *iris.Application

func Start() {
	go startHttpServer()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := app.Shutdown(ctx)
	if err != nil {
		gl.ERR("auth shutdown failed: ", err)
	}
}

/*
 * Http Server
 * serve client request: register|login|loginByGuest
 */
func startHttpServer() {
	app = iris.New()

	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
	app.Use(recover.New())
	app.Use(logger.New())

	registerHandlers(app)

	err := app.Run(iris.Addr(":" + gosconf.AUTH_SERVICE_PORT))

	if err != nil && err != http.ErrServerClosed {
		gl.ERR("Start auth service failed: ", err)
		panic(err)
	}
}

func registerHandlers(app *iris.Application) {
	app.Get("/healthz", k8sHealthHandler)
	app.Post("/register", registerHandler)
	app.Post("/login", loginHandler)
	app.Post("/loginByGuest", loginByGuestHandler)
}

func k8sHealthHandler(ctx iris.Context) {
	_, _ = ctx.Text("ok")
}

func registerHandler(ctx iris.Context) {
	username := ctx.PostValue("username")
	password := ctx.PostValue("password")

	gl.INFO("username: ", username, " password: ", password)

	user, err := account_utils.Lookup(username)
	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user != nil {
		gl.INFO("user exist: ", user.Username)
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_username_already_used"})
		return
	}

	gl.INFO("HandleRegister, username: ", username)
	user, err = account.Create(username, password, account.ACCOUNT_NORMAL)

	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_internal_error"})
		return
	}

	dispathAndRsp(ctx, user)
}

func loginHandler(ctx iris.Context) {
	username := ctx.PostValue("username")
	password := ctx.PostValue("password")

	user, err := account_utils.Lookup(username)
	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user == nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_user_not_found"})
		return
	}

	if !account.Auth(user, password) {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_password_invalid"})
		return
	}

	dispathAndRsp(ctx, user)
}

/*
 * Guest login without register
 */
func loginByGuestHandler(ctx iris.Context) {
	username := ctx.PostValue("username")

	user, err := account_utils.Lookup(username)
	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user == nil {
		user, err = account.Create(username, username, account.ACCOUNT_GUEST)
	}

	if user.Category != account.ACCOUNT_GUEST {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_password_invalid"})
		return
	}

	dispathAndRsp(ctx, user)
}

/*
 * Private Methods
 */
func dispathAndRsp(ctx iris.Context, user *account_utils.Account) {
	session, err := session_utils.Find(user.Uuid)
	if err == nil && session.GameAppId != "" && session.SceneId != "" {
		if game, err := game_utils.Find(session.GameAppId); err == nil {
			loginRsp(ctx, game.Host, game.Port, user.Uuid, session.Token, session.SceneId)
			return
		}
	}

	host, port, session, err := account.Dispatch(user)

	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": "error_internal_error"})
		return
	}

	loginRsp(ctx, host, port, user.Uuid, session.Token, session.SceneId)
}

func loginRsp(ctx iris.Context, host, port, accountId, token, sceneId string) {
	_, _ = ctx.JSON(iris.Map{
		"status":       "success",
		"connectHost":  host,
		"port":         port,
		"accountId":    accountId,
		"sessionToken": token,
		"sceneId":      sceneId})
}
