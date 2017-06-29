package session

import (
	"sync/atomic"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"go.uber.org/zap"
)

func (s *Session) pubCounter() uint64 {
	return atomic.AddUint64(&s._pubCounter, 1)
}

// SendPublish sends the msg to the client
func (s *Session) SendPublish(msg *packets.PublishPacket) {
	if msg.Qos > 0 {
		msg.MessageID = uint16(s.pubCounter())
		s.pendingMu.Lock()
		s.pendingPub = s.pendingPub.Insert(msg)
		if PublishQueueLimit != 0 && s.pendingPub.Len() > PublishQueueLimit {
			s.pendingPub = s.pendingPub[s.pendingPub.Len()-PublishQueueLimit:]
		}
		s.pendingMu.Unlock()
	}
	if s.CanSubscribeTo(msg.TopicName) && s.inFlight() < InFlightLimit && s.send(msg) {
		switch msg.Qos {
		case 1:
			s.pendingMu.Lock()
			s.pendingPub = s.pendingPub.Remove(msg.MessageID)
			s.pendingAck = s.pendingAck.Insert(msg)
			s.pendingMu.Unlock()
		case 2:
			s.pendingMu.Lock()
			s.pendingPub = s.pendingPub.Remove(msg.MessageID)
			s.pendingRec = s.pendingRec.Insert(msg)
			s.pendingMu.Unlock()
		}
	}
}

// ReceivePublish receives the msg from the client
func (s *Session) ReceivePublish(msg *packets.PublishPacket) {
	var dup bool
	if msg.Qos == 2 {
		s.pendingMu.Lock()
		if s.pendingRel.Index(msg.MessageID) != -1 {
			dup = true // already published this message
		}
		s.pendingMu.Unlock()
	}
	if !dup && s.CanPublishTo(msg.TopicName) {
		s.log.Debug("publish", zap.String("topic", msg.TopicName), zap.Int("size", len(msg.Payload)))
		s.deliver(msg)
	}
	switch msg.Qos {
	case 0:
	case 1:
		s.SendPuback(msg.MessageID)
	case 2:
		s.SendPubrec(msg.MessageID)
	}
}

// SendPuback sends a Puback to the client
func (s *Session) SendPuback(msgID uint16) {
	puback := packets.NewControlPacket(packets.Puback)
	puback.(*packets.PubackPacket).MessageID = msgID
	s.send(puback)
}

// ReceivePuback receives the msg from the client
func (s *Session) ReceivePuback(msg *packets.PubackPacket) {
	s.pendingMu.Lock()
	s.pendingAck = s.pendingAck.Remove(msg.MessageID)
	s.pendingMu.Unlock()
}

// SendPubrec sends a Pubrec to the client
func (s *Session) SendPubrec(msgID uint16) {
	pubrec := packets.NewControlPacket(packets.Pubrec)
	pubrec.(*packets.PubrecPacket).MessageID = msgID
	s.pendingMu.Lock()
	s.pendingRel = s.pendingRel.Insert(pubrec)
	s.pendingMu.Unlock()
	s.send(pubrec)
}

// ReceivePubrec receives the msg from the client
func (s *Session) ReceivePubrec(msg *packets.PubrecPacket) {
	s.pendingMu.Lock()
	s.pendingRec = s.pendingRec.Remove(msg.MessageID)
	s.pendingMu.Unlock()
	s.SendPubrel(msg.MessageID)
}

// SendPubrel sends a Pubrel to the client
func (s *Session) SendPubrel(msgID uint16) {
	pubrel := packets.NewControlPacket(packets.Pubrel)
	pubrel.(*packets.PubrelPacket).MessageID = msgID
	s.pendingMu.Lock()
	s.pendingComp = s.pendingComp.Insert(pubrel)
	s.pendingMu.Unlock()
	s.send(pubrel)
}

// ReceivePubrel receives the msg from the client
func (s *Session) ReceivePubrel(msg *packets.PubrelPacket) {
	s.pendingMu.Lock()
	s.pendingRel = s.pendingRel.Remove(msg.MessageID)
	s.pendingMu.Unlock()
	s.SendPubcomp(msg.MessageID)
}

// SendPubcomp sends a Pubcomp to the client
func (s *Session) SendPubcomp(msgID uint16) {
	pubComp := packets.NewControlPacket(packets.Pubcomp)
	pubComp.(*packets.PubcompPacket).MessageID = msgID
	s.send(pubComp)
}

// ReceivePubcomp receives the msg from the client
func (s *Session) ReceivePubcomp(msg *packets.PubcompPacket) {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	s.pendingComp = s.pendingComp.Remove(msg.MessageID)
}

// ResendPending re-sends all pending messages
func (s *Session) ResendPending() {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	for _, msg := range s.pendingComp { // re-send all PUBREL packets that have not been PUBCOMPed
		msg.(*packets.PubrelPacket).Dup = true
		s.send(msg)
	}
	for _, msg := range s.pendingRec { // re-send all PUBLISH packets that have not been PUBRECed
		msg.(*packets.PublishPacket).Dup = true
		s.send(msg)
	}
	for _, msg := range s.pendingAck { // re-send all PUBLISH packets that have not been PUBACKed
		msg.(*packets.PublishPacket).Dup = true
		s.send(msg)
	}
	for _, msg := range s.pendingPub { // send all PUBLISH packets that have not been sent before
		s.send(msg)
	}
}
