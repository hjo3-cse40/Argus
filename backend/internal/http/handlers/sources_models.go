package handlers

import (
	"fmt"
	"strings"
	"time"

	"argus-backend/internal/store"
)

// CreateSourceRequest represents the incoming payload for creating a source
type CreateSourceRequest struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	RepositoryURL  string `json:"repository_url,omitempty"`
	DiscordWebhook string `json:"discord_webhook"`
	WebhookSecret  string `json:"webhook_secret,omitempty"`
}

// SourceResponse represents the outgoing source (excludes webhook_secret)
type SourceResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	RepositoryURL  string    `json:"repository_url,omitempty"`
	DiscordWebhook string    `json:"discord_webhook"`
	CreatedAt      time.Time `json:"created_at"`
}

// ErrorResponse represents error messages
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

// Validate checks if the CreateSourceRequest has valid data
func (r *CreateSourceRequest) Validate() error {
	var details []string

	// Validate name
	if strings.TrimSpace(r.Name) == "" {
		details = append(details, "name is required")
	} else if len(r.Name) > 100 {
		details = append(details, "name must be 100 characters or less")
	}

	// Validate type
	validTypes := map[string]bool{
		"github":  true,
		"gitlab":  true,
		"generic": true,
	}
	if strings.TrimSpace(r.Type) == "" {
		details = append(details, "type is required")
	} else if !validTypes[r.Type] {
		details = append(details, "type must be one of: github, gitlab, generic")
	}

	// Validate Discord webhook URL
	if strings.TrimSpace(r.DiscordWebhook) == "" {
		details = append(details, "discord_webhook is required")
	} else if !strings.HasPrefix(r.DiscordWebhook, "https://discord.com/api/webhooks/") {
		details = append(details, "discord_webhook must be a valid Discord webhook URL")
	}

	// Validate webhook secret length
	if len(r.WebhookSecret) > 256 {
		details = append(details, "webhook_secret must be 256 characters or less")
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// ValidationError represents validation failures
type ValidationError struct {
	Details []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", strings.Join(e.Details, ", "))
}

// toSourceResponse converts a store.Source to SourceResponse (excludes secret)
func toSourceResponse(s store.Source) SourceResponse {
	return SourceResponse{
		ID:             s.ID,
		Name:           s.Name,
		Type:           s.Type,
		RepositoryURL:  s.RepositoryURL,
		DiscordWebhook: s.DiscordWebhook,
		CreatedAt:      s.CreatedAt,
	}
}

// toStoreSource converts CreateSourceRequest to store.Source
func (r *CreateSourceRequest) toStoreSource() store.Source {
	return store.Source{
		Name:           r.Name,
		Type:           r.Type,
		RepositoryURL:  r.RepositoryURL,
		DiscordWebhook: r.DiscordWebhook,
		WebhookSecret:  r.WebhookSecret,
	}
}
