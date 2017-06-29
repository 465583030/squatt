package session

import (
	"github.com/htdvisser/pkg/store"
	"github.com/htdvisser/pkg/store/stringmap"
)

// Store for sessions
type Store struct {
	store store.Interface
}

// NewStore returns a new session store
func NewStore() *Store {
	return &Store{store: stringmap.New()}
}

// New creates a new Session, deleting an old one if existed
func (s *Store) New(name string) *Session {
	session := NewSession(name)
	oldI, existed := s.store.Store(name, session)
	if existed {
		oldI.(*Session).Delete()
	}
	return session
}

// GetOrNew creates a new Session, but returns an old one if existed
func (s *Store) GetOrNew(name string) (*Session, bool) {
	sessionI, existed := s.store.LoadOrBuild(name, func() interface{} {
		return NewSession(name)
	})
	return sessionI.(*Session), existed
}

// Delete a session
func (s *Store) Delete(name string) {
	sessionI, ok := s.store.Delete(name)
	if ok {
		sessionI.(*Session).Delete()
	}
}
