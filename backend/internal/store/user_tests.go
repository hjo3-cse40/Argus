package store

import (
	"strings"
	"testing"
)

// Test CreateUser creates user and returns no error
func TestCreateUser_Success(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}

	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	retrieved, found := store.GetUserByEmail("test@example.com")
	if !found {
		t.Fatal("Expected to find created user")
	}

	if retrieved.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrieved.Email)
	}
	if retrieved.PasswordHash != "hashedpassword123" {
		t.Errorf("Expected password hash 'hashedpassword123', got '%s'", retrieved.PasswordHash)
	}
	if retrieved.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if retrieved.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

// Test CreateUser with duplicate email returns error
func TestCreateUser_DuplicateEmail(t *testing.T) {
	store := NewMemoryStore(100)

	user1 := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user1)
	if err != nil {
		t.Fatalf("First CreateUser failed: %v", err)
	}

	user2 := User{
		Email:        "test@example.com",
		PasswordHash: "differenthash456",
	}
	err = store.CreateUser(user2)
	if err == nil {
		t.Fatal("Second CreateUser should have failed due to duplicate email")
	}

	if !strings.Contains(err.Error(), "email already exists") {
		t.Errorf("Expected duplicate email error, got: %v", err)
	}
}

// Test GetUserByEmail retrieves correct user
func TestGetUserByEmail_Success(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	retrieved, found := store.GetUserByEmail("test@example.com")
	if !found {
		t.Fatal("Expected to find user by email")
	}

	if retrieved.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrieved.Email)
	}
}

// Test GetUserByEmail with non-existent email
func TestGetUserByEmail_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetUserByEmail("nonexistent@example.com")
	if found {
		t.Error("Expected not to find user with non-existent email")
	}
}

// Test GetUserByID retrieves correct user
func TestGetUserByID_Success(t *testing.T) {
	store := NewMemoryStore(100)

	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
	}
	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Get the created user to retrieve its generated ID
	created, found := store.GetUserByEmail("test@example.com")
	if !found {
		t.Fatal("Expected to find created user")
	}

	retrieved, found := store.GetUserByID(created.ID)
	if !found {
		t.Fatal("Expected to find user by ID")
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
	if retrieved.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrieved.Email)
	}
}

// Test GetUserByID with non-existent ID
func TestGetUserByID_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetUserByID("non-existent-id")
	if found {
		t.Error("Expected not to find user with non-existent ID")
	}
}
