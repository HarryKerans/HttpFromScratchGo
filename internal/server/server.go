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
	fmt.Print("HTTP/1.1 200 OK\nContent-Type: text/plain\nContent-Length: 13\nHello World!")
	conn.Close()
}
