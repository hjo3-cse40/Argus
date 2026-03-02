package store

import (
	"testing"
	"time"
)

// Test CreateSession creates session and returns no error
func TestCreateSession_Success(t *testing.T) {
	store := NewMemoryStore(100)

	// Create a user first (sessions reference users)
	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	created, _ := store.GetUserByEmail("test@example.com")

	session := Session{
		UserID:    created.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}

	err = store.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
}

// Test GetSession retrieves correct session
func TestGetSession_Success(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	created, _ := store.GetUserByEmail("test@example.com")

	session := Session{
		UserID:    created.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	err = store.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// We need the generated session ID — create with explicit ID for this test
	sessionWithID := Session{
		SessionID: "test-session-id",
		UserID:    created.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	err = store.CreateSession(sessionWithID)
	if err != nil {
		t.Fatalf("CreateSession with ID failed: %v", err)
	}

	retrieved, found := store.GetSession("test-session-id")
	if !found {
		t.Fatal("Expected to find session by ID")
	}

	if retrieved.SessionID != "test-session-id" {
		t.Errorf("Expected session ID 'test-session-id', got '%s'", retrieved.SessionID)
	}
	if retrieved.UserID != created.ID {
		t.Errorf("Expected user ID '%s', got '%s'", created.ID, retrieved.UserID)
	}
}

// Test GetSession with non-existent session ID
func TestGetSession_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetSession("non-existent-session-id")
	if found {
		t.Error("Expected not to find session with non-existent ID")
	}
}

// Test DeleteSession removes session
func TestDeleteSession_Success(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	created, _ := store.GetUserByEmail("test@example.com")

	session := Session{
		SessionID: "test-session-id",
		UserID:    created.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	err = store.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session exists
	_, found := store.GetSession("test-session-id")
	if !found {
		t.Fatal("Expected session to exist before deletion")
	}

	// Delete session
	err = store.DeleteSession("test-session-id")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify session is gone
	_, found = store.GetSession("test-session-id")
	if found {
		t.Error("Expected session to be deleted")
	}
}

// Test DeleteSession with non-existent ID is idempotent
func TestDeleteSession_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	err := store.DeleteSession("non-existent-session-id")
	if err != nil {
		t.Errorf("Expected no error for non-existent session ID, got: %v", err)
	}
}

// Test DeleteExpiredSessions removes only expired sessions
func TestDeleteExpiredSessions_RemovesOnlyExpired(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	created, _ := store.GetUserByEmail("test@example.com")

	// Create an active session (expires in the future)
	activeSession := Session{
		SessionID: "active-session",
		UserID:    created.ID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	err = store.CreateSession(activeSession)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Create an expired session (expired in the past)
	expiredSession := Session{
		SessionID: "expired-session",
		UserID:    created.ID,
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
		ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
	}
	err = store.CreateSession(expiredSession)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Delete expired sessions
	err = store.DeleteExpiredSessions()
	if err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}

	// Verify active session still exists
	_, found := store.GetSession("active-session")
	if !found {
		t.Error("Expected active session to still exist")
	}

	// Verify expired session is gone
	_, found = store.GetSession("expired-session")
	if found {
		t.Error("Expected expired session to be deleted")
	}
}
