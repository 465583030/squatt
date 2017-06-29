package server

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/session"
	"github.com/htdvisser/squatt/topic"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSubscription(t *testing.T) {
	Convey(`Given a Subcription`, t, func() {
		fooSession := session.NewSession("foo")
		fooTopic := topic.NewTopic("foo")
		s := NewSubscription(fooSession, fooTopic, 1)
		ch := make(chan packets.ControlPacket, 1)
		fooSession.Connect(ch)

		Convey(`When delivering a message`, func() {
			msg := new(packets.PublishPacket)
			msg.Qos = 2
			s.Deliver(msg)
			Convey(`Then the message should be delivered to the session`, func() { So(ch, ShouldNotBeEmpty) })
			Convey(`Then the QoS should be downgraded to 1`, func() { So((<-ch).(*packets.PublishPacket).Qos, ShouldEqual, 1) })
		})
	})

	Convey(`Testing server subscriptions`, t, func() {
		s := NewServer()

		fooSession := session.NewSession("foo")
		fooTopic := topic.NewTopic("foo")

		fooSub := s.Subscribe(fooSession, fooTopic, 1)
		So(s.TopicSubscriptions(fooTopic), ShouldContain, fooSub)
		So(s.SessionSubscriptions(fooSession), ShouldContain, fooSub)

		s.Subscribe(fooSession, fooTopic, 2)
		So(fooSub.qos.Load(), ShouldEqual, 2)

		barTopic := topic.NewTopic("bar")

		barSession := session.NewSession("bar")
		s.Unsubscribe(barSession, barTopic)
		s.Subscribe(barSession, barTopic, 0)

		barSub := s.Subscribe(fooSession, barTopic, 2)
		So(s.TopicSubscriptions(fooTopic), ShouldNotContain, barSub)
		So(s.TopicSubscriptions(barTopic), ShouldContain, barSub)
		So(s.SessionSubscriptions(fooSession), ShouldContain, barSub)

		s.Unsubscribe(fooSession, fooTopic)
		So(s.TopicSubscriptions(fooTopic), ShouldNotContain, fooSub)
		So(s.SessionSubscriptions(fooSession), ShouldNotContain, fooSub)

		s.Unsubscribe(fooSession)

		s.Unsubscribe(fooSession, barTopic)
	})
}
