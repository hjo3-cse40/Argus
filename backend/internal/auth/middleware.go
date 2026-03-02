package auth

import (
	"argus-backend/internal/store"
	"context"
	"net/http"
	"time"
)

// contextKey is an unexported type for context keys in this package
type contextKey string

const userContextKey contextKey = "user"

// RequireAuth is middleware that protects routes requiring a valid session
// It validates the session cookie and injects the user into the request context
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.validateSession(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Inject user into request context for downstream handlers
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateSession checks that the session cookie corresponds to a valid, non-expired session
func (s *Service) validateSession(r *http.Request) (store.User, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return store.User{}, ErrMissingSessionCookie
	}

	session, found := s.store.GetSession(cookie.Value)
	if !found {
		return store.User{}, ErrSessionNotFound
	}

	if isExpired(session) {
		// Clean up expired session
		_ = s.store.DeleteSession(cookie.Value)
		return store.User{}, ErrSessionExpired
	}

	user, found := s.store.GetUserByID(session.UserID)
	if !found {
		return store.User{}, ErrUserNotFound
	}

	return user, nil
}

// UserFromContext retrieves the authenticated user from the request context
// Returns false if no user is present (i.e. route was not protected by RequireAuth)
func UserFromContext(ctx context.Context) (store.User, bool) {
	user, ok := ctx.Value(userContextKey).(store.User)
	return user, ok
}

// isExpired checks if a session has passed its expiry time
func isExpired(session store.Session) bool {
	return time.Now().UTC().After(session.ExpiresAt)
}
