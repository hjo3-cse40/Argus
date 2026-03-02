package store

import (
	"time"

	"github.com/google/uuid"
)

// CreateSession stores a new session
func (s *MemoryStore) CreateSession(session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session.SessionID == "" {
		session.SessionID = uuid.New().String()
	}

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}

	s.sessions = append(s.sessions, session)
	return nil
}

// GetSession retrieves a session by session ID
func (s *MemoryStore) GetSession(sessionID string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sess := range s.sessions {
		if sess.SessionID == sessionID {
			return sess, true
		}
	}
	return Session{}, false
}

// DeleteSession removes a session by session ID (logout)
func (s *MemoryStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sess := range s.sessions {
		if sess.SessionID == sessionID {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return nil
		}
	}
	return nil // not found is not an error — idempotent
}

// DeleteExpiredSessions removes all sessions past their expiry time
func (s *MemoryStore) DeleteExpiredSessions() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	active := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		if sess.ExpiresAt.After(now) {
			active = append(active, sess)
		}
	}
	s.sessions = active
	return nil
}
