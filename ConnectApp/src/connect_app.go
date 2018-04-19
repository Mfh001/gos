package main

import (
	"net"
	"log"
	"time"
	"io"
	"encoding/binary"
)

const TCP_TIMEOUT = 6

/*
 * 连接服务
 *
 * 连接校验
 * 消息转发
 * 广播管理
 */

func main() {
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
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	header := make([]byte, 2)

	for {
		conn.SetReadDeadline(time.Now().Add(TCP_TIMEOUT * time.Second))
		n, err := io.ReadFull(conn, header)
		if err != nil {
			NOTICE("Connection Closed: ", err)
			break
		}

		size := binary.BigEndian.Uint16(header)
		data := make([]byte, size)
		n, err = io.ReadFull(conn, data)
		if err != nil {
			WARN("error receiving msg, bytes: ", n, "reason: ", err)
			break
		}
	}
}