package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	s := &Server{}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s.listener = listener
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}
	s.closed.Store(true)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n" + // Status line
		"Content-Type: text/plain\r\n" + // Example header
		"Content-Length: 13\r\n" + // Content length header
		"\r\n" + // Blank line to separate headers from the body
		"Hello World!\n" // Body
	conn.Write([]byte(response))
	return
}
