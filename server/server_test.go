package server

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

type mockClient struct {
	in        io.Writer
	queue     chan packets.ControlPacket
	out       io.Reader
	responses []packets.ControlPacket
}

func newMockClient() *mockClient {
	c := &mockClient{
		queue: make(chan packets.ControlPacket),
	}

	out, outPipe := io.Pipe()
	c.out = out
	go func() {
		for msg := range c.queue {
			msg.Write(outPipe)
		}
		outPipe.Close()
	}()

	in, inPipe := io.Pipe()
	c.in = inPipe
	go func() {
		for {
			pkt, err := packets.ReadPacket(in)
			if err != nil {
				return
			}
			c.responses = append(c.responses, pkt)
		}
	}()
	return c
}

func (c *mockClient) Read(p []byte) (n int, err error) {
	return c.out.Read(p)
}

func (c *mockClient) Write(p []byte) (n int, err error) {
	return c.in.Write(p)
}

func (c *mockClient) Close() {
	close(c.queue)
}

func (c *mockClient) Send(msg packets.ControlPacket) {
	c.queue <- msg
}

func (c *mockClient) Responses() []packets.ControlPacket {
	return c.responses
}

func TestServer(t *testing.T) {
	Convey(`Given a Server and mock Client`, t, func() {
		s := NewServer()

		log, _ := zap.NewDevelopment()
		s.SetLogger(log)

		newConnect := func() *packets.ConnectPacket {
			connect := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
			connect.ProtocolName = "MQTT"
			connect.ProtocolVersion = 0x04
			connect.CleanSession = true
			return connect
		}

		disconnect := packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)

		run := func(commands ...packets.ControlPacket) (responses []packets.ControlPacket, err error) {
			c := newMockClient()
			resCh := make(chan error)
			go func() {
				resCh <- s.NewClient().handle(c)
			}()
			for _, command := range commands {
				c.Send(command)
			}
			c.Close()
			err = <-resCh
			time.Sleep(10 * time.Millisecond)
			responses = c.Responses()
			return
		}

		Convey(`When sending a CONNECT packet`, func() {
			responses, err := run(newConnect())
			Convey(`Then there should be a positive response`, func() {
				So(responses, ShouldHaveLength, 1)
				So(responses[0], ShouldHaveSameTypeAs, new(packets.ConnackPacket))
				So(responses[0].(*packets.ConnackPacket).ReturnCode, ShouldEqual, packets.Accepted)
			})
			Convey(`Then the connection should be closed`, func() { So(err, ShouldEqual, io.EOF) })
		})

		Convey(`When sending a CONNECT packet with invalid protocol`, func() {
			invalidConnect := newConnect()
			invalidConnect.ProtocolName = "invalid"
			responses, err := run(invalidConnect)
			Convey(`Then the server should return a protocol violation error`, func() { So(err.Error(), ShouldContainSubstring, "Protocol Violation") })
			Convey(`Then there should be a negative response`, func() {
				So(responses, ShouldHaveLength, 1)
				So(responses[0], ShouldHaveSameTypeAs, new(packets.ConnackPacket))
				So(responses[0].(*packets.ConnackPacket).ReturnCode, ShouldEqual, packets.ErrProtocolViolation)
			})
		})

		Convey(`When sending multiple CONNECT packets`, func() {
			_, err := run(newConnect(), newConnect())
			Convey(`Then the server should return a protocol violation error`, func() { So(err.Error(), ShouldContainSubstring, "Protocol Violation") })
		})

		Convey(`When sending a CONNECT + DISCONNECT`, func() {
			responses, err := run(newConnect(), disconnect)
			Convey(`Then there should be a positive response`, func() {
				So(responses, ShouldHaveLength, 1)
				So(responses[0], ShouldHaveSameTypeAs, new(packets.ConnackPacket))
				So(responses[0].(*packets.ConnackPacket).ReturnCode, ShouldEqual, packets.Accepted)
			})
			Convey(`Then the connection should be closed`, func() { So(err, ShouldEqual, io.EOF) })
		})

		// TODO: Test other packet types

	})

	Convey(`Given a Server listening on a TCP port`, t, func() {
		s := NewServer()
		lis, err := net.Listen("tcp", ":0")
		_, port, _ := net.SplitHostPort(lis.Addr().String())
		if err != nil {
			panic(err)
		}
		go s.Serve(lis)

		Convey(`When connecting to the server`, func() {
			conn, err := net.Dial("tcp", "localhost:"+port)
			if err != nil {
				panic(err)
			}
			time.Sleep(10 * time.Millisecond)
			Convey(`Then the server should have 1 connection`, func() { So(s.stats.sockets, ShouldEqual, 1) })
			Convey(`When disconnecting the client`, func() {
				conn.Close()
				Convey(`Then the server should have 0 connections`, func() { So(s.stats.sockets, ShouldEqual, 0) })
			})
		})
		Convey(`When trying to listen on a busy port`, func() {
			err := s.ListenAndServe(":" + port)
			Convey(`Then an error should be returned`, func() { So(err, ShouldNotBeNil) })
		})
		Convey(`When trying to listen on a port that is not busy`, func() {
			lis.Close()
			Convey(`Then no error should be returned`, func() {
				errCh := make(chan error, 1)
				go func() {
					errCh <- s.ListenAndServe(":" + port)
				}()
				select {
				case err := <-errCh:
					So(err, ShouldNotBeNil)
				case <-time.After(10 * time.Millisecond):
				}
			})
		})
	})
}
