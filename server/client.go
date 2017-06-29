package server

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/session"
	"go.uber.org/zap"
)

// Buffer sizes
var (
	ClientSendBufferSize = 16
)

// Client connection
type Client struct {
	server *Server
	log    *zap.Logger

	session   *session.Session
	keepAlive *watchdog

	sendCh chan packets.ControlPacket

	ctx    context.Context
	cancel context.CancelFunc

	errMu sync.Mutex
	err   error
}

type wrappedErr struct {
	err error
}

func (c *Client) setError(err error) {
	if err == nil {
		return
	}
	c.errMu.Lock()
	if c.err == nil {
		c.err = err
	}
	c.errMu.Unlock()
	c.cancel()
	c.session.Disconnect()
}

func (c *Client) getError() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	return c.err
}

// NewClient creates a new MQTT Client
func (s *Server) NewClient() *Client {
	c := &Client{
		server: s,
		log:    s.log,
		sendCh: make(chan packets.ControlPacket),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

// Handle the client connection
func (c *Client) Handle(conn net.Conn) error {
	c.log = c.log.With(zap.String("addr", conn.RemoteAddr().String()))
	return c.handle(conn)
}

func (c *Client) handle(rw io.ReadWriter) error {
	waitSend := make(chan struct{})
	go func() {
		c.sendRoutine(rw)
		close(waitSend)
	}()
	go c.receiveRoutine(rw)
	<-c.ctx.Done()
	if err := c.getError(); err == nil {
		c.setError(c.ctx.Err())
	}
	close(c.sendCh)
	c.keepAlive.Stop()
	<-waitSend
	return c.getError().(error)
}
