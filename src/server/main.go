package main

import (
	"encoding/binary"
	"fmt"
	"gen_server"
	"io"
	"log"
	"manager"
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

	gen_server.StartNamingServer()
	time.Sleep(1 * time.Second)
	gen_server.Start("root_manager", new(manager.RootManager), "root_manager")

	start := time.Now()
	count := 1
	for i := 0; i < count; i++ {
		gen_server.Call("root_manager", "SystemInfo", "root_manager", 2014)
		gen_server.Cast("root_manager", "SystemInfo", "root_manager", 2014)
	}
	fmt.Println("duration: ", time.Now().Sub(start).Seconds())
	fmt.Println("Per Second: ", int(float64(count)/time.Now().Sub(start).Seconds()))

	// gen_server.Stop("root_manager", "Say goodbye!")
	fmt.Println("Server Started!")
	start_tcp_server()
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
