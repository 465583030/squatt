package server

import (
	"sync/atomic"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/session"
	"github.com/htdvisser/squatt/topic"
)

// Subscription of session->topic with a qos
type Subscription struct {
	topic   *topic.Topic
	session *session.Session
	qos     atomic.Value
}

// NewSubscription returns a new Subscription
// Subscriptions should usually be created with the server.Subscribe func
func NewSubscription(session *session.Session, topic *topic.Topic, qos byte) *Subscription {
	sub := &Subscription{session: session, topic: topic}
	sub.qos.Store(qos)
	return sub
}

// Deliver a copy of msg to the subscription
func (s *Subscription) Deliver(msg *packets.PublishPacket) {
	publish := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	publish.TopicName = msg.TopicName
	publish.Payload = msg.Payload
	publish.Qos = msg.Qos
	if qos := s.qos.Load().(byte); qos < publish.Qos {
		publish.Qos = qos
	}
	s.session.SendPublish(publish)
}
