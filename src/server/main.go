package main

import (
	"api"
	"encoding/binary"
	"fmt"
	"gen_server"
	"io"
	"log"
	// "manager"
	"net"
	"runtime"
	"time"
	. "utils"
)

func main() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("caught panic in main()", x)
		}
	}()

	// runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(1)

	// go SysRoutine()

	duckAge := 1
	duck := func(words string) {
		fmt.Println("words is: ", words)
		fmt.Println("duckAge is: ", duckAge)
	}

	duck("hello")

	gen_server.StartNamingServer()
	time.Sleep(1 * time.Second)
	fmt.Println("Server Started!")

	server_name := "test_server"
	gen_server.Start(server_name, new(Player), server_name)
	ret := ""
	gen_server.Call(server_name, "Wrap", func() {
		ret = "hello"
	})

	fmt.Println("ret: ", ret)

	// masureDynamic(1000000)
	start_tcp_server()
}

func masureDynamic(count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		CallWithArgs(new(api.Encoder), "EncodeEquip", "", "", "")
		CallWithArgs(new(api.Encoder), "EncodeEquip", "", "", "")
		CallWithArgs(new(api.Encoder), "EncodeEquip", "", "", "")
		// encoder := new(api.Encoder)
		// encoder.EncodeEquip("", "", "")
	}
	seconds := time.Now().Sub(start).Seconds()
	fmt.Println("duration: ", seconds)
	fmt.Println("Per Second: ", int(float64(count)/seconds))
}

func start_tcp_server() {
	l, err := net.Listen("tcp", ":4100")
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
	defer func() {
		if x := recover(); x != nil {
			ERR("caught panic in handleClient", x)
		}
	}()

	server_name := conn.RemoteAddr().String()
	player_server := gen_server.Start(server_name, new(Player), server_name)
	// response_channel := make(chan []byte)

	header := make([]byte, 2)
	bufctrl := make(chan bool)

	defer func() {
		close(bufctrl)
	}()

	// create write buffer
	out := NewBuffer(conn, bufctrl)
	go out.Start()

	for {
		// header
		conn.SetReadDeadline(time.Now().Add(TCP_TIMEOUT * time.Second))
		n, err := io.ReadFull(conn, header)
		if err != nil {
			NOTICE("Connection Closed: ", err)
			break
		}

		// data
		size := binary.BigEndian.Uint16(header)
		data := make([]byte, size)
		n, err = io.ReadFull(conn, data)
		if err != nil {
			WARN("error receiving msg, bytes:", n, "reason:", err)
			break
		}

		// gen_server.Cast(server_name, "HandleRequest", data, out)
		player_server.Cast("HandleRequest", data, out)
	}

}
