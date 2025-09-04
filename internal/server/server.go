package server

import (
	"fmt"
	"httpfromtcp/internal/response"
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
			fmt.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response.WriteStatusLine(conn, response.StatusCodeSuccess)
	headers := response.GetDefaultHeaders(0)
	err := response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Printf("error writing response, %s", err)
	}
	return
}
