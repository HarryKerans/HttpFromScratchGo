package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
			linesChan := getLineslinesannel(c)
			for line := range linesChan {
				fmt.Println(line)
			}
			c.Close()
			fmt.Println("Connection to", conn.RemoteAddr(), "Closed!")
		}(conn)

	}

}

func getLineslinesannel(tcpLine io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer tcpLine.Close()
		defer close(lines)
		currentLine := ""
		for {
			buffer := make([]byte, 8)
			n, err := tcpLine.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLine != "" {
						lines <- currentLine
						currentLine = ""
					}
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			chunk := buffer[:n]
			parts := bytes.Split(chunk, []byte("\n"))

			currentLine += string(parts[0])
			if len(parts) != 1 {
				lines <- currentLine
				currentLine = string(parts[len(parts)-1])
			}
		}
	}()
	return lines
}
