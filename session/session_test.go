package session

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSession(t *testing.T) {
	Convey(`Given a Session`, t, func() {
		s := NewSession("foo")
		Convey(`Then that session should have a name`, func() { So(s.Name(), ShouldEqual, "foo") })
		Convey(`When setting the session to persistent`, func() {
			s.SetPersistent()
			Convey(`Then the session should be persistent`, func() { So(s.Persistent(), ShouldBeTrue) })
		})
		Convey(`When sending a control packet`, func() {
			res := s.send(&packets.PublishPacket{})
			Convey(`Then the result should be negative (there is no client channel)`, func() { So(res, ShouldBeFalse) })
		})
		Convey(`When setting the client channel`, func() {
			ch := make(chan packets.ControlPacket, 1)
			s.Connect(ch)
			Convey(`When setting the client channel again`, func() {
				s.Connect(make(chan packets.ControlPacket))
				Convey(`Then the old channel should be closed`, func() {
					_, ok := <-ch
					So(ok, ShouldBeFalse)
				})
			})
			Convey(`When sending a control packet`, func() {
				msg := &packets.PublishPacket{TopicName: "foo"}
				res := s.send(msg)
				Convey(`Then the result should be positive`, func() { So(res, ShouldBeTrue) })
				Convey(`Then the client channel should contain that control packet`, func() {
					So(ch, ShouldNotBeEmpty)
					So(<-ch, ShouldEqual, msg)
				})
				Convey(`When sending another control packet`, func() {
					res := s.send(&packets.PublishPacket{})
					Convey(`Then the result should be negative (the channel is full)`, func() { So(res, ShouldBeFalse) })
				})
			})
			Convey(`When disconnecting the session`, func() {
				var onDisconnectCalled bool
				s.SetOnDisconnect(func() {
					onDisconnectCalled = true
				})
				s.Disconnect()
				Convey(`Then the client channel should be closed`, func() {
					_, ok := <-ch
					So(ok, ShouldBeFalse)
				})
				Convey(`Then the OnDisconnect func should have been called`, func() {
					So(onDisconnectCalled, ShouldBeTrue)
				})
			})
		})
		Convey(`When publishing a publish packet`, func() {
			res := s.deliver(&packets.PublishPacket{})
			Convey(`Then the result should be negative (there is no publish channel)`, func() { So(res, ShouldBeFalse) })
		})
		Convey(`When setting the publish channel`, func() {
			ch := make(chan *packets.PublishPacket, 1)
			s.DeliverTo(ch)
			Convey(`When publishing a publish packet`, func() {
				msg := &packets.PublishPacket{TopicName: "foo"}
				res := s.deliver(msg)
				Convey(`Then the result should be positive`, func() { So(res, ShouldBeTrue) })
				Convey(`Then the publish channel should contain that publish packet`, func() {
					So(ch, ShouldNotBeEmpty)
					So(<-ch, ShouldEqual, msg)
				})
				Convey(`When publishing another publish packet`, func() {
					res := s.deliver(&packets.PublishPacket{})
					Convey(`Then the result should be negative (the channel is full)`, func() { So(res, ShouldBeFalse) })
				})
			})
		})
		Convey(`When setting the session will on a connected client`, func() {
			s.Connect(make(chan packets.ControlPacket, 1))
			ch := make(chan *packets.PublishPacket, 1)
			s.DeliverTo(ch)
			s.SetWill("foo", []byte("bar"), 1, true)
			Convey(`When disconnecting`, func() {
				s.Disconnect()
				Convey(`Then the will should have been published`, func() {
					So(ch, ShouldNotBeEmpty)
				})
			})
			Convey(`When clearing the session will`, func() {
				s.ClearWill()
				Convey(`When disconnecting`, func() {
					s.Disconnect()
					Convey(`Then no will has been published`, func() {
						So(ch, ShouldBeEmpty)
					})
				})
			})
		})
		Convey(`When deleting the session`, func() {
			var onDeleteCalled bool
			s.SetOnDelete(func() {
				onDeleteCalled = true
			})
			s.Delete()
			Convey(`Then the OnDelete func should have been called`, func() {
				So(onDeleteCalled, ShouldBeTrue)
			})
		})
	})
}
