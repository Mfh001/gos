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

package api_server

import (
	"context"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	gl "github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/player"
	"net/http"
	"time"
)

var app *iris.Application

func StartApiServer() {
	go startHttpServer()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := app.Shutdown(ctx)
	if err != nil {
		gl.ERR("api mgr shutdown failed: ", err)
	}
}

func startHttpServer() {
	app = iris.New()
	app.Use(recover.New())
	app.Use(logger.New())

	registerHandlers(app)

	err := app.Run(iris.Addr("127.0.0.1:4100"))

	if err != nil && err != http.ErrServerClosed {
		gl.ERR("Start api mgr service failed: ", err)
		panic(err)
	}
}

func registerHandlers(app *iris.Application) {
	app.Post("/dump_player", dumpPlayerHandler)
}

func dumpPlayerHandler(ctx iris.Context) {
	playerId := ctx.PostValue("playerId")
	data, err := player.GetCtx(playerId)
	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": err.Error(),
		})
		return
	}
	_, _ = ctx.JSON(data)
}
