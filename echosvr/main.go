package main

import (
	"fmt"
	"net"
	"time"
)

const (
	PORT = "8080"
	respHeader = "HTTP/1.1 200 OK\r\nContent-Type: text/html;\n\n"
)

func main() {
	// Init the server
	server, err := net.Listen("tcp", ":" +PORT)
	if server == nil {
		fmt.Printf("Fail to start listening: %v", err)
	}
	defer server.Close()

	// Loop for listen and handle..
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Printf("Fail to Accept %v", err)
		}

		// start a goroutine to handle new connection
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// Resp header
	_, err := conn.Write([]byte(respHeader))
	if err != nil {
		fmt.Printf("Fail to write resp: %v", err)
	}

	// Body is the same as request (including request header).
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	for {
		buf := make([]byte, 512)
		_, err := conn.Read(buf)
		if err != nil { // EOF, timeout or worse
			return
		}

		fmt.Printf("From:%s\n%s", conn.RemoteAddr().String(), string(buf))
		_, err = conn.Write(buf)
		if err != nil {
			fmt.Printf("Fail to write resp: %v", err)
		}
	}
}
