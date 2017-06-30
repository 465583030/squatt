// Package server impelents an MQTT Server
package server

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/auth"
	"github.com/htdvisser/squatt/session"
	"github.com/htdvisser/squatt/topic"
	"go.uber.org/zap"
)

type serverStats struct {
	// BEGIN sync/atomic aligned
	sockets int64
	// END sync/atomic aligned
}

// Server implements an MQTT Server
type Server struct {
	log   *zap.Logger
	stats *serverStats

	auth     auth.Plugin
	sessions *session.Store
	topics   *topic.Store

	subscriptionsMu      sync.RWMutex
	sessionSubscriptions map[*session.Session]subscriptionsByTopic
	topicSubscriptions   map[*topic.Topic]subscriptionsBySession

	retainedMessagesMu sync.RWMutex
	retainedMessages   map[*topic.Topic]*packets.PublishPacket

	publish chan *packets.PublishPacket
}

// NewServer returns a new MQTT Server
func NewServer() *Server {
	s := &Server{
		log:   zap.NewNop(),
		stats: new(serverStats),

		auth:     auth.NoAuth,
		sessions: session.NewStore(),
		topics:   topic.NewStore(),

		sessionSubscriptions: make(map[*session.Session]subscriptionsByTopic),
		topicSubscriptions:   make(map[*topic.Topic]subscriptionsBySession),

		retainedMessages: make(map[*topic.Topic]*packets.PublishPacket),

		publish: make(chan *packets.PublishPacket, 512),
	}

	return s
}

// Route publish messages. Calling this from multiple goroutines increases parallellism
func (s *Server) Route() {
	for msg := range s.publish {
		if msg.Retain {
			s.RetainMessage(msg)
		}
		topics := s.topics.Match(msg.TopicName)
		subscriptions := s.TopicSubscriptions(topics...)
		s.log.Info(
			"publish",
			zap.String("topic", msg.TopicName),
			zap.Int("matching-topics", len(topics)),
			zap.Int("matching-subscriptions", len(subscriptions)),
		)
		for _, sub := range subscriptions {
			sub.Deliver(msg)
		}
	}
}

// SetLogger sets the logger on the server
func (s *Server) SetLogger(log *zap.Logger) {
	s.log = log
}

// ListenAndServe on an address
func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Debug("server listening", zap.String("addr", lis.Addr().String()))
	return s.Serve(lis)
}

// ListenAndServeTLS is similar to ListenAndServe, except that it uses TLS
func (s *Server) ListenAndServeTLS(addr string, config *tls.Config) error {
	lis, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	s.log.Debug("tls server listening", zap.String("addr", lis.Addr().String()))
	return s.Serve(lis)
}

// Serve on the given listener
func (s *Server) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer conn.Close()
			conns := atomic.AddInt64(&s.stats.sockets, 1)
			s.log.Debug("accept connection", zap.String("addr", conn.RemoteAddr().String()), zap.Int64("conns", conns))
			s.NewClient().Handle(conn)
			conns = atomic.AddInt64(&s.stats.sockets, -1)
			s.log.Debug("release connection", zap.String("addr", conn.RemoteAddr().String()), zap.Int64("conns", conns), zap.Error(err))
		}()
	}
}
