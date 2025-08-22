package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	host := "127.0.0.1:42069"

	fmt.Println("Program Start")
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		log.Fatal(err)
	}

	conn, connerr := net.DialUDP("udp", nil, addr)
	if connerr != nil {
		log.Fatal(connerr)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			log.Fatal(readErr)
		}
		_, writeErr := conn.Write(line)
		if writeErr != nil {
			log.Fatal(writeErr)
		}

	}
}
