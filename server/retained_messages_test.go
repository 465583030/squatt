package server

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRetainedMessages(t *testing.T) {
	Convey(`Given a Server`, t, func() {
		s := NewServer()

		pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
		pub.TopicName = "foo"
		pub.Retain = true
		pub.Payload = []byte("foo")

		Convey(`When getting the retained messages`, func() {
			msgs := s.RetainedMessages(s.topics.Match("#")...)
			Convey(`Then there should be no messages`, func() { So(msgs, ShouldBeEmpty) })
		})
		Convey(`When trying to retain a message without the retain bit`, func() {
			pub := pub
			pub.Retain = false
			s.RetainMessage(pub)
			Convey(`When getting the retained messages`, func() {
				msgs := s.RetainedMessages(s.topics.Match("#")...)
				Convey(`Then there should be no messages`, func() { So(msgs, ShouldBeEmpty) })
			})
		})
		Convey(`When retaining a message`, func() {
			pub := pub
			s.RetainMessage(pub)
			Convey(`When getting the retained messages`, func() {
				msgs := s.RetainedMessages(s.topics.Match("#")...)
				Convey(`Then the retained message should be returned`, func() { So(msgs, ShouldContain, pub) })
			})
			Convey(`When retaining a message without payload`, func() {
				pub := pub
				pub.Payload = nil
				s.RetainMessage(pub)
				Convey(`When getting the retained messages`, func() {
					msgs := s.RetainedMessages(s.topics.Match("#")...)
					Convey(`Then there should be no messages`, func() { So(msgs, ShouldBeEmpty) })
				})
			})
		})
	})
}
