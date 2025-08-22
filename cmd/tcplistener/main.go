package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {

	listener, listenError := net.Listen("tcp4", "127.0.0.1:42069")
	if listenError != nil {
		log.Fatal(listenError)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		fmt.Println("Connection Accepted From", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			request, err := request.RequestFromReader(c)
			if err != nil {
				log.Fatalf("error parsing request: %s\n", err.Error())
			}
			fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", request.RequestLine.Method, request.RequestLine.RequestTarget, request.RequestLine.HttpVersion)
			c.Close()
			fmt.Println("Connection to", conn.RemoteAddr(), "Closed!")
		}(conn)

	}

}
