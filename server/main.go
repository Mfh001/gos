package main

import (
	"bytes"
	"fmt"
	"gen_server"
	"log"
	"manager"
	"net"
	"runtime"
	"time"
)

func main() {
	// defer func() {
	// 	if x := recover(); x != nil {
	// 		fmt.Println("caught panic in main()", x)
	// 	}
	// }()
	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1)

	gen_server.Start("root_manager", new(manager.RootManager), "root_manager")
	gen_server.Start("naming_server", new(gen_server.NamingServer))

	start := time.Now()
	count := 1
	for i := 0; i < count; i++ {
		gen_server.Call("root_manager", "SystemInfo", "root_manager", 2014)
		gen_server.Cast("root_manager", "SystemInfo", "root_manager", 2014)
	}
	fmt.Println("duration: ", time.Now().Sub(start).Seconds())
	fmt.Println("Per Second: ", int(float64(count)/time.Now().Sub(start).Seconds()))

	// gen_server.Cast("root_manager", "wrap", params, nil)
	// gen_server.Stop("root_manager", "Say goodbye!")
	fmt.Println("Server Started!")
	start_tcp_server()
}

func start_tcp_server() {
	// Listen on TCP port 4100 on all interfaces.
	l, err := net.Listen("tcp", ":4100")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	server_name := conn.RemoteAddr().String()
	gen_server.Start(server_name, new(manager.RootManager), server_name)
	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			conn.Close()
			// go gen_server.Stop(server_name, "Connection closed!")
			break
		}

		// Send a response back to person contacting us.
		n := bytes.Index(buf, []byte{0})
		s := string(buf[:n])
		result, _ := gen_server.Call(server_name, "Echo", s)
		fmt.Println("result: ", result[0].String())
		// conn.Write([]byte(result[0].String()))

		// gen_server.Call(server_name, "Echo", s)
		conn.Write([]byte(s))
	}
}
