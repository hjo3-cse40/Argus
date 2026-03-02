package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateUser adds a new user to the store
func (s *MemoryStore) CreateUser(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate email
	for _, u := range s.users {
		if u.Email == user.Email {
			return &ValidationError{Details: []string{fmt.Sprintf("email already exists: %s", user.Email)}}
		}
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	s.users = append(s.users, user)
	return nil
}

// GetUserByEmail retrieves a user by email address
func (s *MemoryStore) GetUserByEmail(email string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return u, true
		}
	}
	return User{}, false
}

// GetUserByID retrieves a user by ID
func (s *MemoryStore) GetUserByID(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == id {
			return u, true
		}
	}
	return User{}, false
}
