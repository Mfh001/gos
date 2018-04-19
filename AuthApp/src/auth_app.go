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
	"github.com/go-redis/redis"
	"github.com/grpc/grpc-go"
	"log"
	"context"
	"time"
	pb "connectAppProto"
)

const (
	SERVER_PORT = "3000"
	REDIS_URL = "localhost:6379"
	REDIS_PASSWORD = ""
	REDIS_DB = 0

	GRPC_CONNECT_APP_ADDR = "localhost:50051"
)

type ConnectAppInfo struct {
	host string
	port string
}

type Session struct {
	accountId string
	serverId string
	token string
	connect *ConnectAppInfo
}

var redisClient *redis.Client

func main() {
	connectDB()
	startHttpServer()
}

func connectDB() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     REDIS_URL,
		Password: REDIS_PASSWORD, // no password set
		DB:       REDIS_DB,  // use default DB
	})
}

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
	account := ctx.Params().Get("account")
	password := ctx.Params().Get("password")

	ctx.JSON(iris.Map{"message": "Hello iris web framework."})
}

func loginHandler(ctx iris.Context) {
	account := ctx.Params().Get("account")
	password := ctx.Params().Get("password")

	ctx.JSON(iris.Map{"message": "Hello iris web framework."})
}

func loginByGuestHandler(ctx iris.Context) {
	account := ctx.Params().Get("account")
	password := ctx.Params().Get("password")

	session := &Session{
		accountId: ""
	}

	chooseConnectApp(session)

	ctx.JSON(iris.Map{"message": "Hello iris web framework."})
}

func chooseConnectApp(session *Session) {
	conn, err := grpc.Dial(GRPC_CONNECT_APP_ADDR, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDispatcherClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.DispatchPlayer(ctx, &pb.DispatchRequest{AccountId:session.accountId, ServerId:session.serverId})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %s:%s", r.GetConnectAppHost(), r.GetConnectAppPort())
}
