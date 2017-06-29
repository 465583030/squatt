package session

import (
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/auth"
	"go.uber.org/zap"
)

var (
	// PublishQueueLimit limits the amount of messages that can be queued for publishing to a client
	// If this limit is exceeded, old messages are dropped
	PublishQueueLimit = 32

	// InFlightLimit limits the amount of messages that can be unacknowledged by a client.
	// While this limit is exceeded, no more publish messages are sent to this client
	InFlightLimit = 32
)

// NewSession returns a new session with the given name
func NewSession(name string) *Session {
	s := &Session{name: name}
	s.initialize()
	return s
}

// Session for MQTT Client
type Session struct {
	// BEGIN sync/atomic aligned
	_pubCounter uint64 // access with pubCounter()
	// END sync/atomic aligned

	// BEGIN unprotected - must not be changed after initialization
	name         string
	auth         auth.Interface
	onDisconnect func()
	onDelete     func()
	log          *zap.Logger
	persistent   bool
	deliveryCh   chan<- *packets.PublishPacket
	// END unprotected

	// BEGIN mu protected
	mu    sync.Mutex
	will  *packets.PublishPacket
	outCh chan<- packets.ControlPacket
	// END mu protected

	// BEGIN pendingMu protected
	pendingMu   sync.Mutex
	pendingPub  pendingMessages // []*PublishPacket
	pendingAck  pendingMessages // []*PublishPacket
	pendingRec  pendingMessages // []*PublishPacket
	pendingRel  pendingMessages // []*PubrecPacket
	pendingComp pendingMessages // []*PubrelPacket
	// END pendingMu protected
}

// initialize all fields to their zero value
func (s *Session) initialize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	s._pubCounter = 0
	s.auth, _ = auth.NoAuth(s.name, "", nil)
	s.onDisconnect = func() {}
	s.onDelete = func() {}
	s.log = zap.NewNop()
	s.persistent = false
	s.deliveryCh = nil
	s.will = nil
	s.outCh = nil
	s.pendingPub = make(pendingMessages, 0, PublishQueueLimit)
	s.pendingAck = make(pendingMessages, 0, InFlightLimit)
	s.pendingRec = make(pendingMessages, 0, InFlightLimit)
	s.pendingRel = make(pendingMessages, 0, InFlightLimit)
	s.pendingComp = make(pendingMessages, 0, InFlightLimit)
}

// Name of the session
func (s *Session) Name() string {
	return s.name
}

// SetAuth sets the authentication for this session
func (s *Session) SetAuth(auth auth.Interface) {
	s.auth = auth
}

// CanPublishTo returns true if the session can publish to the given topic
func (s *Session) CanPublishTo(topic string) bool {
	return s.auth.CanPublishTo(topic)
}

// CanSubscribeTo returns true if the session can subscribe to the given topic
func (s *Session) CanSubscribeTo(topic string) bool {
	return s.auth.CanSubscribeTo(topic)
}

// SetOnDisconnect sets the function that is executed on disconnection of the session
func (s *Session) SetOnDisconnect(onDisconnect func()) {
	s.onDisconnect = onDisconnect
}

// SetOnDelete sets the function that is executed on deletion of the session
func (s *Session) SetOnDelete(onDelete func()) {
	s.onDelete = onDelete
}

// SetLogger sets the logger for this session
func (s *Session) SetLogger(log *zap.Logger) {
	s.log = log.With(zap.String("id", s.name))
}

// SetPersistent sets the session persistency
func (s *Session) SetPersistent() {
	s.persistent = true
}

// Persistent returns the session persistency
func (s *Session) Persistent() bool {
	return s.persistent
}

// SetWill sets the session will
func (s *Session) SetWill(topic string, payload []byte, qos uint8, retain bool) {
	if !s.CanPublishTo(topic) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Debug("set will", zap.String("topic", topic))
	s.will = &packets.PublishPacket{TopicName: topic, Payload: payload}
	s.will.Qos, s.will.Retain = qos, retain
}

// ClearWill clears the session will
func (s *Session) ClearWill() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.will != nil {
		s.log.Debug("clear will")
		s.will = nil
	}
}

func (s *Session) publishWill() {
	if s.will == nil {
		return
	}
	s.log.Debug("publish will", zap.String("topic", s.will.TopicName))
	s.deliver(s.will)
	s.will = nil
}

// DeliverTo sets the channel that should be used to publish packets to the server
func (s *Session) DeliverTo(ch chan<- *packets.PublishPacket) {
	s.deliveryCh = ch
}

// deliver a message to the application
func (s *Session) deliver(msg *packets.PublishPacket) bool {
	if s.deliveryCh == nil {
		return false
	}
	select {
	case s.deliveryCh <- msg:
		return true
	default:
	}
	return false
}

// Connect connects the session to a client
func (s *Session) Connect(ch chan<- packets.ControlPacket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.outCh != nil {
		s.log.Debug("disconnect old connection")
		close(s.outCh)
		s.outCh = ch
		s.publishWill()

		s.mu.Unlock()
		s.onDisconnect() // onDisconnect should be called without lock
		s.mu.Lock()
	} else {
		s.outCh = ch
	}
	s.log.Debug("connect")
}

// Disconnect the session
func (s *Session) Disconnect() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.outCh == nil {
		return
	}
	s.log.Debug("disconnect")
	close(s.outCh)
	s.outCh = nil
	s.publishWill()

	s.mu.Unlock()
	s.onDisconnect() // onDisconnect should be called without lock
	s.mu.Lock()
}

// send a control packet to the client
func (s *Session) send(msg packets.ControlPacket) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.outCh == nil {
		return false
	}
	select {
	case s.outCh <- msg:
		return true
	default:
	}
	return false
}

// Delete the session
func (s *Session) Delete() {
	if s == nil {
		return
	}
	s.Disconnect()
	s.log.Debug("delete")
	s.onDelete()
	s.initialize() // re-initialize the session for re-use
}
