// Package auth contains authentication for the MQTT Server
package auth

// Interface for authentication
type Interface interface {
	Username() string
	CanConnect() bool
	CanPublishTo(topic string) bool
	CanSubscribeTo(topic string) bool
}

// Plugin for authentication
type Plugin func(clientIdentifier string, username string, password []byte) (Interface, error)

// NoAuth does not restrict
func NoAuth(clientIdentifier string, username string, password []byte) (Interface, error) {
	return &noAuth{username: username}, nil
}

type noAuth struct {
	username string
}

func (n noAuth) Username() string                 { return n.username }
func (n noAuth) CanConnect() bool                 { return true }
func (n noAuth) CanPublishTo(topic string) bool   { return true }
func (n noAuth) CanSubscribeTo(topic string) bool { return true }
