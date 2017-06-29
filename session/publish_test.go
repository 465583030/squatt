package session

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPublish(t *testing.T) {
	Convey(`Given a Session`, t, func() {
		s := NewSession("foo")
		msg := packets.PublishPacket{TopicName: "foo", Payload: []byte("foo")}
		Convey(`When sending a QoS 0 Publish Message`, func() {
			msg := msg
			s.SendPublish(&msg)
			Convey(`Then it should not be in the publish queue`, func() { So(s.pendingPub, ShouldBeEmpty) })
		})
		Convey(`When sending a QoS 1 Publish Message`, func() {
			msg := msg
			msg.Qos = 1
			s.SendPublish(&msg)
			Convey(`Then it should have a message ID`, func() { So(msg.MessageID, ShouldNotEqual, 0) })
			Convey(`Then it should be in the publish queue`, func() { So(s.pendingPub, ShouldNotBeEmpty) })
		})
		Convey(`When sending a QoS 2 Publish Message`, func() {
			msg := msg
			msg.Qos = 2
			s.SendPublish(&msg)
			Convey(`Then it should have a message ID`, func() { So(msg.MessageID, ShouldNotEqual, 0) })
			Convey(`Then it should be in the publish queue`, func() { So(s.pendingPub, ShouldNotBeEmpty) })
		})
		Convey(`When the session has a client channel`, func() {
			ch := make(chan packets.ControlPacket, 1)
			s.Connect(ch)
			Convey(`When sending a QoS 0 Publish Message`, func() {
				msg := msg
				s.SendPublish(&msg)
				Convey(`Then it should not be in the publish queue`, func() { So(s.pendingPub, ShouldBeEmpty) })
				Convey(`Then it should be in the client channel`, func() { So(ch, ShouldNotBeEmpty); So(<-ch, ShouldEqual, &msg) })
			})
			Convey(`When sending a QoS 1 Publish Message`, func() {
				msg := msg
				msg.Qos = 1
				s.SendPublish(&msg)
				Convey(`Then it should have a message ID`, func() { So(msg.MessageID, ShouldNotEqual, 0) })
				Convey(`Then it should be in the puback queue`, func() { So(s.pendingAck, ShouldNotBeEmpty) })
				Convey(`Then it should be in the client channel`, func() { So(ch, ShouldNotBeEmpty); So(<-ch, ShouldEqual, &msg) })
				Convey(`When receiving a Puback Message`, func() {
					s.ReceivePuback(&packets.PubackPacket{MessageID: msg.MessageID})
					Convey(`Then it should no longer be in the puback queue`, func() { So(s.pendingAck, ShouldBeEmpty) })
				})
			})
			Convey(`When sending a QoS 2 Publish Message`, func() {
				msg := msg
				msg.Qos = 2
				s.SendPublish(&msg)
				Convey(`Then it should have a message ID`, func() { So(msg.MessageID, ShouldNotEqual, 0) })
				Convey(`Then it should be in the pubrec queue`, func() { So(s.pendingRec, ShouldNotBeEmpty) })
				Convey(`Then it should be in the client channel`, func() { So(ch, ShouldNotBeEmpty); So(<-ch, ShouldEqual, &msg) })
				Convey(`When receiving a Pubrec Message`, func() {
					<-ch
					s.ReceivePubrec(&packets.PubrecPacket{MessageID: msg.MessageID})
					Convey(`Then it should no longer be in the pubrec queue`, func() { So(s.pendingRec, ShouldBeEmpty) })
					Convey(`Then a pubrel should be in the pubcomp queue`, func() { So(s.pendingComp, ShouldNotBeEmpty) })
					Convey(`Then a pubrel should be in the client channel`, func() {
						So(ch, ShouldNotBeEmpty)
						pubrel := <-ch
						So(pubrel, ShouldHaveSameTypeAs, &packets.PubrelPacket{})
						So(pubrel.Details().MessageID, ShouldEqual, msg.MessageID)
					})
					Convey(`When receiving a Pubcomp Message`, func() {
						<-ch
						s.ReceivePubcomp(&packets.PubcompPacket{MessageID: msg.MessageID})
						Convey(`Then it should no longer be in the pubcomp queue`, func() { So(s.pendingComp, ShouldBeEmpty) })
					})
				})
			})
			Convey(`When receiving a QoS 0 Publish Message`, func() {
				msg := msg
				s.ReceivePublish(&msg)
				Convey(`Then no response should be in the client channel`, func() { So(ch, ShouldBeEmpty) })
			})
			Convey(`When receiving a QoS 1 Publish Message`, func() {
				msg := msg
				msg.Qos = 1
				msg.MessageID = 1
				s.ReceivePublish(&msg)
				Convey(`Then a puback should be in the client channel`, func() {
					So(ch, ShouldNotBeEmpty)
					puback := <-ch
					So(puback, ShouldHaveSameTypeAs, &packets.PubackPacket{})
					So(puback.Details().MessageID, ShouldEqual, msg.MessageID)
				})
			})
			Convey(`When receiving a QoS 2 Publish Message`, func() {
				msg := msg
				msg.Qos = 2
				msg.MessageID = 1
				s.ReceivePublish(&msg)
				Convey(`Then it should be in the pubrel queue`, func() { So(s.pendingRel, ShouldNotBeEmpty) })
				Convey(`Then a pubrec should be in the client channel`, func() {
					So(ch, ShouldNotBeEmpty)
					pubrec := <-ch
					So(pubrec, ShouldHaveSameTypeAs, &packets.PubrecPacket{})
					So(pubrec.Details().MessageID, ShouldEqual, msg.MessageID)
				})
				Convey(`When receiving a Pubrel Message`, func() {
					<-ch
					s.ReceivePubrel(&packets.PubrelPacket{MessageID: msg.MessageID})
					Convey(`Then it should no longer be in the pubrel queue`, func() { So(s.pendingRel, ShouldBeEmpty) })
				})
			})
		})
		Convey(`When there are pending messages`, func() {
			pendingPub := &packets.PublishPacket{MessageID: 1}
			s.pendingPub = s.pendingPub.Insert(pendingPub)
			pendingAck := &packets.PublishPacket{MessageID: 2}
			s.pendingAck = s.pendingAck.Insert(pendingAck)
			pendingRec := &packets.PublishPacket{MessageID: 3}
			s.pendingRec = s.pendingRec.Insert(pendingRec)
			pendingRel := &packets.PubrecPacket{MessageID: 4}
			s.pendingRel = s.pendingRel.Insert(pendingRel)
			pendingComp := &packets.PubrelPacket{MessageID: 5}
			s.pendingComp = s.pendingComp.Insert(pendingComp)
			Convey(`Then the non-pendingPub should count towards the inFlight`, func() { So(s.inFlight(), ShouldEqual, 4) })
			Convey(`When calling ResendPending`, func() {
				ch := make(chan packets.ControlPacket, 10)
				s.Connect(ch)
				s.ResendPending()
				// Server MUST re-send any unacknowledged PUBLISH Packets (where QoS > 0) and PUBREL Packets using their original Packet Identifiers [MQTT-4.4.0-1]
				Convey(`Then the messages should be sent to the client channel`, func() {
					So(ch, ShouldHaveLength, 4)
					So(<-ch, ShouldEqual, pendingComp)
					So(pendingComp.Dup, ShouldBeTrue)
					So(<-ch, ShouldEqual, pendingRec)
					So(pendingRec.Dup, ShouldBeTrue)
					So(<-ch, ShouldEqual, pendingAck)
					So(pendingAck.Dup, ShouldBeTrue)
					So(<-ch, ShouldEqual, pendingPub)
					So(pendingPub.Dup, ShouldBeFalse)
				})
			})
		})
	})
}
