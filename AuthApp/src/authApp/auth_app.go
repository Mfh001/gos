/*
验证服务器

功能：
	1.账户注册
	2.登录验证
	3.连接服分配
*/

package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"google.golang.org/grpc"
	"account"
	pb "connectAppProto"
	"goslib/redisDB"
	gl "goslib/logger"
	"log"
)

const (
	SERVER_PORT = "3000"
	REDIS_URL = "localhost:6379"
	REDIS_PASSWORD = ""
	REDIS_DB = 0

	GRPC_CONNECT_APP_ADDR = "localhost:50051"
)

type Session struct {
	accountId string
	token string
	connectHost string
	connectPort string
}

func main() {
	redisDB.Connect(REDIS_URL, REDIS_PASSWORD, REDIS_DB)
	connectConnectApp()
	startHttpServer()
}

/*
 * connect to ConnectAppMgr
 */
func connectConnectApp() {
	conn, err := grpc.Dial(GRPC_CONNECT_APP_ADDR, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	account.ConnectRpcClient = pb.NewDispatcherClient(conn)
}

/*
 * Http Server
 * serve client request: register|login|loginByGuest
 */
func startHttpServer() {
	app := iris.New()

	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
	app.Use(recover.New())
	app.Use(logger.New())

	registerHandlers(app)

	app.Run(iris.Addr(":" + SERVER_PORT))
}

func registerHandlers(app *iris.Application) {
	app.Post("/register", registerHandler)
	app.Post("/login", loginHandler)
	app.Post("/loginByGuest", loginByGuestHandler)
}

func registerHandler(ctx iris.Context) {
	username := ctx.PostValue("username")
	password := ctx.PostValue("password")

	user, err := account.Lookup(username)
	if err != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_username_already_used"})
		return
	}

	gl.INFO("HandleRegister, username: ", username)
	user, err = account.Create(username, password)

	if err != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_internal_error"})
		return
	}

	dispathAndRsp(ctx, user)
}

func loginHandler(ctx iris.Context) {
	username := ctx.PostValue("username")
	password := ctx.PostValue("password")

	user, err := account.Lookup(username)
	if err != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user == nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_user_not_found"})
		return
	}

	if !user.Auth(password) {
		ctx.JSON(iris.Map{
			"status": "failed",
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

	user, err := account.Lookup(username)
	if err != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_internal_error"})
		return
	}

	if user == nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_user_not_found"})
		return
	}

	if user.Category != account.ACCOUNT_GUEST {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_password_invalid"})
		return
	}

	dispathAndRsp(ctx, user)
}

/*
 * Private Methods
 */
func dispathAndRsp(ctx iris.Context, user *account.Account) {
	host, port, err := user.Dispatch()

	if err != nil {
		ctx.JSON(iris.Map{
			"status": "failed",
			"error_code": "error_internal_error"})
		return
	}

	ctx.JSON(iris.Map{
		"status": "success",
		"connectHost": host,
		"port": port,
		"accountId": user.Uuid})
}
