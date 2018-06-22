package main

import (
	"net"
	"net/url"

	"github.com/davecgh/go-spew/spew"
	"github.com/eleztian/mqtt-server/transport"
	"github.com/pkg/errors"
	"gopkg.in/tomb.v2"
)

type Server struct {
	conn net.Listener
	tomb *tomb.Tomb
}

type Conn struct {
	listen net.Listener
}

func NewServer(urlStr string) (*Server, error) {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		tomb: &tomb.Tomb{},
	}
	switch u.Scheme {
	case "tcp":
		lst, err := net.Listen("tcp", u.Host)
		if err != nil {
			return nil, err
		}
		server.conn = lst
	default:
		return nil, errors.New("unknown net type " + u.Scheme)
	}
	return server, nil
}

func (s *Server) Start() error {
	for {
		conn, err := s.conn.Accept()
		if err != nil {
			return err
		}
		c := transport.NewConn(conn)
		go handle(c)
	}

	return nil
}

func handle(c *transport.Conn) error {
	p, err := c.Receive()
	if err != nil {
		return err
	}
	spew.Dump(p)
	return nil
}

func main() {
	server, err := NewServer("tcp://0.0.0.0:1883")
	if err != nil {
		panic(err)
	}
	server.Start()
}
