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
package api_mgr

import (
	"context"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/mafei198/gos/goslib/account_utils"
	"github.com/mafei198/gos/goslib/game_utils"
	"github.com/mafei198/gos/goslib/gen/proto"
	"github.com/mafei198/gos/goslib/gosconf"
	gl "github.com/mafei198/gos/goslib/logger"
	"github.com/mafei198/gos/goslib/scene_utils"
	"github.com/mafei198/gos/goslib/utils/redis_utils"
	"github.com/mafei198/gos/world/game_mgr"
	"google.golang.org/grpc"
	"net/http"
	"sync"
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
		gl.ERR("api mgr shutdown failed: ", err)
	}
}

func startHttpServer() {
	app = iris.New()
	app.Use(recover.New())
	app.Use(logger.New())

	registerHandlers(app)

	err := app.Run(iris.Addr("127.0.0.1:3100"))

	if err != nil && err != http.ErrServerClosed {
		gl.ERR("Start api mgr service failed: ", err)
		panic(err)
	}
}

func registerHandlers(app *iris.Application) {
	app.Post("/add_scene", addSceneHandler)
	app.Post("/deploy_scene", deploySceneHandler)
}

func addSceneHandler(ctx iris.Context) {
	name := ctx.PostValue("name")
	category := redis_utils.ToInt32(ctx.PostValue("category"))
	confId := redis_utils.ToInt32(ctx.PostValue("confId"))

	// 创建scene
	scene, err := scene_utils.Add(name, category, confId)
	if err != nil {
		_, _ = ctx.JSON(iris.Map{
			"status":     "failed",
			"error_code": err.Error(),
		})
		return
	}

	// 为scene分配服务器
	info, err := game_mgr.DispatchByScene(scene.Uuid, scene.Uuid)
	if err != nil {
		failRsp(ctx, err)
		return
	}

	// 初始化场景
	deployScene(ctx, scene)

	_, _ = ctx.JSON(iris.Map{
		"status":    "success",
		"uuid":      scene.Uuid,
		"name":      scene.Name,
		"confId":    scene.ConfId,
		"gameAppId": info.AppId})
}

func deploySceneHandler(ctx iris.Context) {
	name := ctx.PostValue("name")
	uuid := scene_utils.GenUuid(name)
	scene, err := scene_utils.Find(uuid)
	if err != nil {
		failRsp(ctx, err)
		return
	}
	deployScene(ctx, scene)
}

func deployScene(ctx iris.Context, scene *scene_utils.Scene) {
	if scene.GameAppId == "" {
		_, err := account_utils.Create(scene.Uuid, scene.Uuid, "", 0)
		if err != nil {
			failRsp(ctx, err)
			return
		}
		info, err := game_mgr.DispatchByScene(scene.Uuid, scene.Uuid)
		if err != nil {
			failRsp(ctx, err)
			return
		}
		scene.GameAppId = info.AppId
	}
	game, err := game_utils.Find(scene.GameAppId)
	if err != nil {
		failRsp(ctx, err)
		return
	}
	client, err := connectGame(game)
	if err != nil {
		failRsp(ctx, err)
		return
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	reply, err := client.DeployScene(timeoutCtx, &proto.DeploySceneRequest{
		SceneId: scene.Uuid,
	})
	if err != nil {
		failRsp(ctx, err)
		return
	}
	_, _ = ctx.JSON(iris.Map{
		"status": reply.Success,
	})
}

var rpcClients = &sync.Map{}

func connectGame(game *game_utils.Game) (proto.GameRpcServerClient, error) {
	if value, ok := rpcClients.Load(game.Uuid); ok {
		return value.(proto.GameRpcServerClient), nil
	}
	conf := gosconf.RPC_FOR_GAME_APP_RPC
	addr := fmt.Sprintf("%s:%s", game.RpcHost, game.RpcPort)
	conn, err := grpc.Dial(addr, conf.DialOptions...)
	if err != nil {
		gl.ERR("connect Game failed: ", err)
		return nil, err
	}
	client := proto.NewGameRpcServerClient(conn)
	rpcClients.Store(game.Uuid, client)
	return client, nil
}

func failRsp(ctx iris.Context, err error) {
	_, _ = ctx.JSON(iris.Map{
		"status":     "failed",
		"error_code": err.Error(),
	})
}
