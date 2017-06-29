package server

import (
	"sort"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/pkg/sortutil"
	"github.com/htdvisser/squatt/session"
	"github.com/htdvisser/squatt/topic"
)

type subscriptionsBySession []*Subscription

func (l subscriptionsBySession) Len() int           { return len(l) }
func (l subscriptionsBySession) Less(i, j int) bool { return l[i].session.Name() < l[j].session.Name() }
func (l subscriptionsBySession) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l subscriptionsBySession) Search(session *session.Session) int {
	return sort.Search(len(l), func(i int) bool {
		return l[i].session.Name() >= session.Name()
	})
}
func (l subscriptionsBySession) Insert(s *Subscription) subscriptionsBySession {
	updated := append(l, s)
	sortutil.SwapLast(updated)
	return updated
}
func (l subscriptionsBySession) Remove(s *Subscription) subscriptionsBySession {
	pos := l.Search(s.session)
	if pos < len(l) && l[pos] == s {
		return append(l[:pos], l[pos+1:]...)
	}
	return l
}

type subscriptionsByTopic []*Subscription

func (l subscriptionsByTopic) Len() int           { return len(l) }
func (l subscriptionsByTopic) Less(i, j int) bool { return l[i].topic.Name() < l[j].topic.Name() }
func (l subscriptionsByTopic) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l subscriptionsByTopic) Search(topic *topic.Topic) int {
	return sort.Search(len(l), func(i int) bool {
		return l[i].topic.Name() >= topic.Name()
	})
}
func (l subscriptionsByTopic) Index(topic *topic.Topic) int {
	pos := l.Search(topic)
	if pos < len(l) && l[pos].topic.Name() == topic.Name() {
		return pos
	}
	return -1
}
func (l subscriptionsByTopic) Load(topic *topic.Topic) (*Subscription, bool) {
	if idx := l.Index(topic); idx != -1 {
		return l[idx], true
	}
	return nil, false
}
func (l subscriptionsByTopic) Insert(s *Subscription) subscriptionsByTopic {
	updated := append(l, s)
	sortutil.SwapLast(updated)
	return updated
}
func (l subscriptionsByTopic) Remove(s *Subscription) subscriptionsByTopic {
	pos := l.Search(s.topic)
	if pos < len(l) && l[pos] == s {
		return append(l[:pos], l[pos+1:]...)
	}
	return l
}

// Subscribe a session to a topic and return the subscription
func (s *Server) Subscribe(session *session.Session, topic *topic.Topic, qos byte) (subscription *Subscription) {
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	sessionSubscriptions, ok := s.sessionSubscriptions[session]
	if ok {
		if subscription, ok = sessionSubscriptions.Load(topic); ok {
			subscription.qos.Store(qos)
			return
		}
	}
	topicSubscriptions, _ := s.topicSubscriptions[topic]
	subscription = NewSubscription(session, topic, qos)

	s.sessionSubscriptions[session] = sessionSubscriptions.Insert(subscription)
	s.topicSubscriptions[topic] = topicSubscriptions.Insert(subscription)

	return
}

// Unsubscribe a session from a topic and return the old subscription, if any
func (s *Server) Unsubscribe(session *session.Session, topic ...*topic.Topic) {
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	sessionSubscriptions, ok := s.sessionSubscriptions[session]
	if !ok {
		return
	}
	if len(topic) == 0 {
		for _, sub := range sessionSubscriptions {
			topic = append(topic, sub.topic)
		}
	}
	for _, topic := range topic {
		subscription, ok := sessionSubscriptions.Load(topic)
		if !ok {
			continue
		}

		sessionSubscriptions = sessionSubscriptions.Remove(subscription)
		if len(sessionSubscriptions) > 0 {
			s.sessionSubscriptions[session] = sessionSubscriptions.Remove(subscription)
		} else {
			delete(s.sessionSubscriptions, session)
		}

		topicSubscriptions, _ := s.topicSubscriptions[topic]
		topicSubscriptions = topicSubscriptions.Remove(subscription)
		if len(topicSubscriptions) > 0 {
			s.topicSubscriptions[topic] = topicSubscriptions
		} else {
			delete(s.topicSubscriptions, topic)
		}
	}

	return
}

// SessionSubscriptions returns all subscriptions of the session
func (s *Server) SessionSubscriptions(session ...*session.Session) (subs []*Subscription) {
	s.subscriptionsMu.RLock()
	defer s.subscriptionsMu.RUnlock()
	for _, session := range session {
		subs = append(subs, s.sessionSubscriptions[session]...)
	}
	return
}

// TopicSubscriptions returns all subscriptions to a topic
func (s *Server) TopicSubscriptions(topic ...*topic.Topic) (subs []*Subscription) {
	s.subscriptionsMu.RLock()
	defer s.subscriptionsMu.RUnlock()
	for _, topic := range topic {
		subs = append(subs, s.topicSubscriptions[topic]...)
	}
	return
}

// Publish returns the publish channel
func (s *Server) Publish() chan<- *packets.PublishPacket {
	return s.publish
}
