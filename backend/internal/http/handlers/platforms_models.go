package handlers

import (
	"strings"
	"time"

	"argus-backend/internal/store"
)

// CreatePlatformRequest represents the incoming payload for creating a platform
type CreatePlatformRequest struct {
	Name                   string `json:"name"`
	DiscordWebhook         string `json:"discord_webhook"`
	WebhookSecret          string `json:"webhook_secret,omitempty"`
	FilterIncludeCombine   string `json:"filter_include_combine,omitempty"`
	FilterExcludeCombine   string `json:"filter_exclude_combine,omitempty"`
}

// UpdatePlatformRequest represents the incoming payload for updating a platform
type UpdatePlatformRequest struct {
	DiscordWebhook         string  `json:"discord_webhook"`
	WebhookSecret          string  `json:"webhook_secret,omitempty"`
	FilterIncludeCombine   *string `json:"filter_include_combine,omitempty"`
	FilterExcludeCombine   *string `json:"filter_exclude_combine,omitempty"`
}

// PlatformResponse represents the outgoing platform (excludes webhook_secret)
type PlatformResponse struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	DiscordWebhook         string    `json:"discord_webhook"`
	FilterIncludeCombine   string    `json:"filter_include_combine"`
	FilterExcludeCombine   string    `json:"filter_exclude_combine"`
	CreatedAt              time.Time `json:"created_at"`
}

// Validate checks if the CreatePlatformRequest has valid data
func (r *CreatePlatformRequest) Validate() *ValidationError {
	var details []string

	// Validate name - must be one of (youtube, reddit, x)
	allowedNames := map[string]bool{
		"youtube": true,
		"reddit":  true,
		"x":       true,
	}
	if strings.TrimSpace(r.Name) == "" {
		details = append(details, "name is required")
	} else if !allowedNames[r.Name] {
		details = append(details, "name must be one of: youtube, reddit, x")
	}

	// Validate discord_webhook only when provided.
	if strings.TrimSpace(r.DiscordWebhook) != "" &&
		!strings.HasPrefix(r.DiscordWebhook, "https://discord.com/api/webhooks/") &&
		!strings.HasPrefix(r.DiscordWebhook, "https://discordapp.com/api/webhooks/") {
		details = append(details, "discord_webhook must start with https://discord.com/api/webhooks/ or https://discordapp.com/api/webhooks/")
	}

	details = append(details, validateCombineFieldsOptional(r.FilterIncludeCombine, r.FilterExcludeCombine)...)

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// Validate checks if the UpdatePlatformRequest has valid data
func (r *UpdatePlatformRequest) Validate() *ValidationError {
	var details []string

	// Validate discord_webhook only when provided.
	if strings.TrimSpace(r.DiscordWebhook) != "" &&
		!strings.HasPrefix(r.DiscordWebhook, "https://discord.com/api/webhooks/") &&
		!strings.HasPrefix(r.DiscordWebhook, "https://discordapp.com/api/webhooks/") {
		details = append(details, "discord_webhook must start with https://discord.com/api/webhooks/ or https://discordapp.com/api/webhooks/")
	}

	if r.FilterIncludeCombine != nil {
		details = append(details, validateCombinePointer(*r.FilterIncludeCombine, "filter_include_combine")...)
	}
	if r.FilterExcludeCombine != nil {
		details = append(details, validateCombinePointer(*r.FilterExcludeCombine, "filter_exclude_combine")...)
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

func validateCombineFieldsOptional(incRaw, excRaw string) []string {
	var details []string
	if strings.TrimSpace(incRaw) != "" {
		v := store.NormalizeFilterCombineString(incRaw)
		if v != "any" && v != "all" {
			details = append(details, "filter_include_combine must be any or all")
		}
	}
	if strings.TrimSpace(excRaw) != "" {
		v := store.NormalizeFilterCombineString(excRaw)
		if v != "any" && v != "all" {
			details = append(details, "filter_exclude_combine must be any or all")
		}
	}
	return details
}

func validateCombinePointer(raw, field string) []string {
	v := store.NormalizeFilterCombineString(raw)
	if v != "any" && v != "all" {
		return []string{field + " must be any or all"}
	}
	return nil
}

// toPlatformResponse converts a store.Platform to PlatformResponse (excludes secret)
func toPlatformResponse(p store.Platform) PlatformResponse {
	return PlatformResponse{
		ID:                   p.ID,
		Name:                 p.Name,
		DiscordWebhook:       p.DiscordWebhook,
		FilterIncludeCombine: store.NormalizeFilterCombineString(p.FilterIncludeCombine),
		FilterExcludeCombine: store.NormalizeFilterCombineString(p.FilterExcludeCombine),
		CreatedAt:            p.CreatedAt,
	}
}

// toStorePlatform converts CreatePlatformRequest to store.Platform
func (r *CreatePlatformRequest) toStorePlatform() store.Platform {
	return store.Platform{
		Name:                   r.Name,
		DiscordWebhook:         r.DiscordWebhook,
		WebhookSecret:          r.WebhookSecret,
		FilterIncludeCombine:   r.FilterIncludeCombine,
		FilterExcludeCombine:   r.FilterExcludeCombine,
	}
}
