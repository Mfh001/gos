package main

import (
	"net"
	"log"
	"time"
	"io"
	"encoding/binary"
	"goslib/logger"
	"account"
	"goslib/packet"
	"api"
	"gslib/routes"
	"connection"
	"goslib/redisDB"
	"agent"
	"google.golang.org/grpc"
	pb "connectAppProto"
)

const GRPC_GAME_APP_MGR_ADDR = "localhost:50052"

/*
 * 连接服务
 *
 * 连接校验
 * 消息转发
 * 广播管理
 */

func main() {
	redisDB.Connect("localhost:6379", "", 0)
	connectGameMgr()
	agent.Setup()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		connection.Start(conn)
	}
}

func connectGameMgr() {
	conn, err := grpc.Dial(GRPC_GAME_APP_MGR_ADDR, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	agent.GameMgrRpcClient = pb.NewGameDispatcherClient(conn)
}

