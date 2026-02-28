package handlers

import (
	"strings"
	"time"

	"argus-backend/internal/store"
)

// CreateSubsourceRequest represents the incoming payload for creating a subsource
type CreateSubsourceRequest struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	URL        string `json:"url,omitempty"`
}

// UpdateSubsourceRequest represents the incoming payload for updating a subsource
type UpdateSubsourceRequest struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	URL        string `json:"url,omitempty"`
}

// SubsourceResponse represents the outgoing subsource with platform information
type SubsourceResponse struct {
	ID           string    `json:"id"`
	PlatformID   string    `json:"platform_id"`
	PlatformName string    `json:"platform_name"`
	Name         string    `json:"name"`
	Identifier   string    `json:"identifier"`
	URL          string    `json:"url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate checks if the CreateSubsourceRequest has valid data
func (r *CreateSubsourceRequest) Validate() *ValidationError {
	var details []string

	// Validate name - must not be empty or whitespace-only
	if strings.TrimSpace(r.Name) == "" {
		details = append(details, "name is required and cannot be empty")
	}

	// Validate identifier - must not be empty or whitespace-only
	if strings.TrimSpace(r.Identifier) == "" {
		details = append(details, "identifier is required and cannot be empty")
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// Validate checks if the UpdateSubsourceRequest has valid data
func (r *UpdateSubsourceRequest) Validate() *ValidationError {
	var details []string

	// Validate name - must not be empty or whitespace-only
	if strings.TrimSpace(r.Name) == "" {
		details = append(details, "name is required and cannot be empty")
	}

	// Validate identifier - must not be empty or whitespace-only
	if strings.TrimSpace(r.Identifier) == "" {
		details = append(details, "identifier is required and cannot be empty")
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// toSubsourceResponse converts a store.SubsourceWithPlatform to SubsourceResponse
func toSubsourceResponse(s store.SubsourceWithPlatform) SubsourceResponse {
	return SubsourceResponse{
		ID:           s.ID,
		PlatformID:   s.PlatformID,
		PlatformName: s.PlatformName,
		Name:         s.Name,
		Identifier:   s.Identifier,
		URL:          s.URL,
		CreatedAt:    s.CreatedAt,
	}
}

// toStoreSubsource converts CreateSubsourceRequest to store.Subsource
func (r *CreateSubsourceRequest) toStoreSubsource(platformID string) store.Subsource {
	return store.Subsource{
		PlatformID: platformID,
		Name:       r.Name,
		Identifier: r.Identifier,
		URL:        r.URL,
	}
}
