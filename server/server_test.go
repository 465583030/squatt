package server

import (
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServer(t *testing.T) {
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
			Convey(`Then the server should have 1 connection`, func() { So(s.stats.connections, ShouldEqual, 1) })
			Convey(`When disconnecting the client`, func() {
				conn.Close()
				Convey(`Then the server should have 0 connections`, func() { So(s.stats.connections, ShouldEqual, 0) })
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
