package handlers

import (
	"strings"
	"time"

	"argus-backend/internal/store"
)

// RegisterRequest is the request body for POST /api/auth/register
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() *store.ValidationError {
	var details []string

	if strings.TrimSpace(r.Email) == "" {
		details = append(details, "email is required")
	}
	if !strings.Contains(r.Email, "@") {
		details = append(details, "email is invalid")
	}
	if strings.TrimSpace(r.Password) == "" {
		details = append(details, "password is required")
	}
	if len(r.Password) < 8 {
		details = append(details, "password must be at least 8 characters")
	}

	if len(details) > 0 {
		return &store.ValidationError{Details: details}
	}
	return nil
}

// LoginRequest is the request body for POST /api/auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() *store.ValidationError {
	var details []string

	if strings.TrimSpace(r.Email) == "" {
		details = append(details, "email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		details = append(details, "password is required")
	}

	if len(details) > 0 {
		return &store.ValidationError{Details: details}
	}
	return nil
}

// UserResponse is the response body for GET /api/auth/me
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
