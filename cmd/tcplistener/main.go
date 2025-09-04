package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
)

func main() {

	listener, listenError := net.Listen("tcp4", "127.0.0.1:42069")
	if listenError != nil {
		fmt.Print(listenError)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		fmt.Println("Connection Accepted From", conn.RemoteAddr())
		if err != nil {
			fmt.Print(err)
		}
		go func(c net.Conn) {
			request, err := request.RequestFromReader(c)
			if err != nil {
				fmt.Printf("error parsing request: %s\n", err.Error())
			}
			fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", request.RequestLine.Method, request.RequestLine.RequestTarget, request.RequestLine.HttpVersion)
			fmt.Println("Headers:")
			for key, value := range request.Headers {
				fmt.Printf("- %s: %s\n", key, value)
			}
			fmt.Printf("Body:\n %s\n", string(request.Body))

			c.Close()
			fmt.Println("Connection to", conn.RemoteAddr(), "Closed!")
		}(conn)

	}

}
