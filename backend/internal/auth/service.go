package auth

import (
	"argus-backend/internal/store"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	sessionCookieName = "session_id"
	sessionDuration   = 24 * time.Hour
)

// Service handles all auth operations
type Service struct {
	store store.Store
}

// NewService creates a new auth service
func NewService(store store.Store) *Service {
	return &Service{store: store}
}

// RegisterUser creates a new user with a hashed password
func (s *Service) RegisterUser(email, password string) error {
	// Check if user already exists
	_, found := s.store.GetUserByEmail(email)
	if found {
		return ErrUserAlreadyExists
	}

	hash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := store.User{
		Email:        email,
		PasswordHash: hash,
	}

	if err := s.store.CreateUser(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// LoginUser validates credentials and creates a session, writing the cookie to the response
func (s *Service) LoginUser(w http.ResponseWriter, email, password string) error {
	user, found := s.store.GetUserByEmail(email)
	if !found {
		return ErrInvalidCredentials
	}

	if err := comparePassword(user.PasswordHash, password); err != nil {
		return ErrInvalidCredentials
	}

	session := store.Session{
		SessionID: uuid.New().String(),
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(sessionDuration),
	}

	if err := s.store.CreateSession(session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Re-fetch to get the generated session ID
	s.setSessionCookie(w, session.SessionID, session.ExpiresAt)
	return nil
}

// Logout deletes the session from the store and clears the cookie
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ErrMissingSessionCookie
	}

	if err := s.store.DeleteSession(cookie.Value); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.clearSessionCookie(w)
	return nil
}

// GetUserFromSession retrieves the user associated with the session cookie on the request
func (s *Service) GetUserFromSession(r *http.Request) (store.User, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return store.User{}, ErrMissingSessionCookie
	}

	session, found := s.store.GetSession(cookie.Value)
	if !found {
		return store.User{}, ErrSessionNotFound
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return store.User{}, ErrSessionExpired
	}

	user, found := s.store.GetUserByID(session.UserID)
	if !found {
		return store.User{}, ErrUserNotFound
	}

	return user, nil
}

// setSessionCookie writes a secure session cookie to the response
func (s *Service) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true, // not accessible via JS
		Secure:   true, // SET TO TRUE FOR HTTPS * IMPORTANT *
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

// clearSessionCookie removes the session cookie from the client
func (s *Service) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}
