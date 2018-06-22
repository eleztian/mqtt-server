package transport

import (
	"sync"
	"net"
	"github.com/eleztian/mqtt-server/packet"
)

type Conn struct {
	stream *packet.Stream
	wMutex  sync.Mutex
	rMutex sync.Mutex
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		stream:packet.NewStream(conn, conn),
	}
}

func (c *Conn) Receive() (packet.Packet, error) {
	c.rMutex.Lock()
	defer c.rMutex.Unlock()

	return c.stream.Read()
}

func (c *Conn) Send(p packet.Packet) error {
	c.wMutex.Lock()
	defer c.wMutex.Unlock()

	return c.stream.Write(p)
}