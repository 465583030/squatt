package topic

import (
	"github.com/htdvisser/pkg/store"
	"github.com/htdvisser/pkg/store/stringmap"
)

// Store for topics
type Store struct {
	store interface {
		store.Interface
		Match(filter string) (values []interface{})
	}
}

// NewStore returns a new topic store
func NewStore() *Store {
	store := stringmap.New()
	store.PathSettings("/", "+", "#")
	return &Store{store: store}
}

// Get a topic
func (s *Store) Get(name string) *Topic {
	topicI, _ := s.store.LoadOrBuild(name, func() interface{} {
		return NewTopic(name)
	})
	return topicI.(*Topic)
}

// Match topics
func (s *Store) Match(filter string) []*Topic {
	topicsI := s.store.Match(filter)
	topics := make([]*Topic, len(topicsI))
	for i, topicI := range topicsI {
		topics[i] = topicI.(*Topic)
	}
	return topics
}
