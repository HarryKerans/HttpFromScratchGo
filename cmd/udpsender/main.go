package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	host := "127.0.0.1:42069"

	fmt.Println("Program Start")
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		fmt.Print(err)
	}

	conn, connerr := net.DialUDP("udp", nil, addr)
	if connerr != nil {
		fmt.Print(connerr)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			fmt.Print(readErr)
		}
		_, writeErr := conn.Write(line)
		if writeErr != nil {
			fmt.Print(writeErr)
		}

	}
}
