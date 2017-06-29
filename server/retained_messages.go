package server

import (
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/topic"
)

// RetainMessage stores a PUBLISH packet if the RETAIN flag is set to 1
func (s *Server) RetainMessage(msg *packets.PublishPacket) {
	if !msg.Retain {
		return
	}
	topic := s.topics.Get(msg.TopicName)
	s.retainedMessagesMu.Lock()
	defer s.retainedMessagesMu.Unlock()
	if len(msg.Payload) > 0 {
		s.retainedMessages[topic] = msg
	} else {
		delete(s.retainedMessages, topic)
	}
}

// RetainedMessages gets all retained PUBLISH packets for the given topics
func (s *Server) RetainedMessages(topics ...*topic.Topic) (msgs []*packets.PublishPacket) {
	s.retainedMessagesMu.RLock()
	defer s.retainedMessagesMu.RUnlock()
	for _, topic := range topics {
		if msg, ok := s.retainedMessages[topic]; ok {
			msgs = append(msgs, msg)
		}
	}
	return
}
