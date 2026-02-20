package svelgo

import (
	"sync"

	"github.com/gorilla/websocket"
)

// PageSession holds the live component tree and WebSocket connection for one
// page instance.
type PageSession struct {
	mu         sync.Mutex
	components map[string]Component
	conn       *websocket.Conn
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*PageSession
}

var globalSessionStore = &sessionStore{
	sessions: make(map[string]*PageSession),
}

func (s *sessionStore) Register(pageID string, components []Component) {
	compMap := make(map[string]Component, len(components))
	for _, c := range components {
		compMap[c.ComponentID()] = c
	}
	s.mu.Lock()
	s.sessions[pageID] = &PageSession{components: compMap}
	s.mu.Unlock()
}

func (s *sessionStore) Get(pageID string) (*PageSession, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[pageID]
	s.mu.RUnlock()
	return sess, ok
}

func (s *sessionStore) Delete(pageID string) {
	s.mu.Lock()
	delete(s.sessions, pageID)
	s.mu.Unlock()
}
