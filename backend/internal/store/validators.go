package store

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidationError represents validation failures
type ValidationError struct {
	Details []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", strings.Join(e.Details, ", "))
}

// validatePlatform checks if a platform has valid data
func validatePlatform(p Platform) error {
	var details []string

	// Validate name - must be one of (youtube, reddit, x)
	allowedNames := map[string]bool{
		"youtube": true,
		"reddit":  true,
		"x":       true,
	}
	if strings.TrimSpace(p.Name) == "" {
		details = append(details, "name is required")
	} else if !allowedNames[p.Name] {
		details = append(details, "name must be one of: youtube, reddit, x")
	}

	// Validate discord_webhook - must start with Discord webhook URL
	if strings.TrimSpace(p.DiscordWebhook) == "" {
		details = append(details, "discord_webhook is required")
	} else if !strings.HasPrefix(p.DiscordWebhook, "https://discord.com/api/webhooks/") && !strings.HasPrefix(p.DiscordWebhook, "https://discordapp.com/api/webhooks/") {
		details = append(details, "discord_webhook must start with https://discord.com/api/webhooks/ or https://discordapp.com/api/webhooks/")
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}

// validateSubsource checks if a subsource has valid data
func validateSubsource(s Subsource) error {
	var details []string

	// Validate name - not empty or whitespace-only
	if strings.TrimSpace(s.Name) == "" {
		details = append(details, "name cannot be empty or whitespace-only")
	}

	// Validate identifier - not empty or whitespace-only
	if strings.TrimSpace(s.Identifier) == "" {
		details = append(details, "identifier cannot be empty or whitespace-only")
	}

	// Validate URL if provided - must be well-formed
	if s.URL != "" {
		parsedURL, err := url.Parse(s.URL)
		if err != nil {
			details = append(details, "url must be well-formed")
		} else if parsedURL.Scheme == "" {
			details = append(details, "url must have a scheme (http or https)")
		}
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}

	return nil
}