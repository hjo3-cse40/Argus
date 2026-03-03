package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"argus-backend/internal/events"
)

type discordPayload struct {
	Username string			`json:"username,omitempty"`
	Embeds	 []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title		string				`json:"title,omitempty"`
	URL			string				`json:"url,omitempty"`
	Description string				`json:"description,omitempty"`
	Footer		*discordEmbedFoot	`json:"footer,omitempty"`
	Timestamp	string				`json:"timestamp,omitempty"`
}

type discordEmbedFoot struct {
	Text string `json:"text,omitempty"`
}

// formatSource extracts hierarchical source information from event metadata
// Returns "Platform - Subsource" if both platform_name and subsource_name are present
// Falls back to event.Source if either is missing
func formatSource(e *events.Event) string {
	platformName, hasPlatform := e.Metadata["platform_name"].(string)
	subsourceName, hasSubsource := e.Metadata["subsource_name"].(string)

	if hasPlatform && hasSubsource && platformName != "" && subsourceName != "" {
		return fmt.Sprintf("%s - %s", platformName, subsourceName)
	}

	// Fallback to legacy source field
	return e.Source
}

// Notifier provides Discord notification functionality with store access
type Notifier struct {
	store Store
}

// Store interface for webhook URL lookup
type Store interface {
	GetSubsource(id string) (Subsource, bool)
	GetPlatform(id string) (Platform, bool)
	GetSourceByName(name string) (Source, bool)
}

// Subsource represents a subsource with platform information
type Subsource struct {
	ID         string
	PlatformID string
	Name       string
}

// Platform represents a platform with webhook configuration
type Platform struct {
	ID             string
	Name           string
	DiscordWebhook string
}

// Source represents a legacy flat source
type Source struct {
	ID             string
	Name           string
	DiscordWebhook string
}

// NewNotifier creates a new Notifier with store access
func NewNotifier(store Store) *Notifier {
	return &Notifier{store: store}
}

// GetWebhookURL retrieves the Discord webhook URL for an event
// Uses hierarchical lookup (subsource -> platform) if subsource_id is present
// Falls back to legacy source lookup for backward compatibility
func (n *Notifier) GetWebhookURL(e *events.Event) (string, error) {
	url, _, err := n.ResolveDestination(e)
	return url, err
}

// ResolveDestination returns the webhook URL and platform ID for an event.
// The platform ID is needed to load per-destination filters.
// Returns ("", "", err) if the destination cannot be resolved.
func (n *Notifier) ResolveDestination(e *events.Event) (webhookURL string, platformID string, err error) {
	subsourceID, ok := e.Metadata["subsource_id"].(string)
	if !ok || subsourceID == "" {
		source, found := n.store.GetSourceByName(e.Source)
		if !found {
			return "", "", fmt.Errorf("source not found: %s", e.Source)
		}
		return source.DiscordWebhook, "", nil
	}

	subsource, found := n.store.GetSubsource(subsourceID)
	if !found {
		return "", "", fmt.Errorf("subsource not found: %s", subsourceID)
	}

	platform, found := n.store.GetPlatform(subsource.PlatformID)
	if !found {
		return "", "", fmt.Errorf("platform not found: %s", subsource.PlatformID)
	}

	return platform.DiscordWebhook, platform.ID, nil
}


func RenderDiscordEmbed(e *events.Event) discordPayload {
	sourceType, _ := e.Metadata["source_type"].(string)
	author, _ := e.Metadata["author"].(string)
	description, _ := e.Metadata["description"].(string)

	var descParts []string
	if sourceType != "" {
		descParts = append(descParts, fmt.Sprintf("**Platform:** %s", sourceType))
	}
	// Use formatSource to display hierarchical source information
	descParts = append(descParts, fmt.Sprintf("**Source:** %s", formatSource(e)))
	if author != "" {
		descParts = append(descParts, fmt.Sprintf("**Author:** %s", author))
	}
	if description != "" {
		descParts = append(descParts, truncate(description, 300))
	}

	desc := strings.Join(descParts, "\n")

	return discordPayload{
		Username: "Argus",
		Embeds: []discordEmbed{
			{
				Title:       truncate(e.Title, 256),
				URL:         e.URL,
				Description: truncate(desc, 4096),
				Footer: &discordEmbedFoot{
					Text: truncate("event_id: "+e.EventID, 2048),
				},
				Timestamp: e.CreatedAt.UTC().Format(time.RFC3339),
			},
		},
	}
}

// Send formatted embed to discord using given webhook URL
func SendDiscordWebhook(webhookURL string, e *events.Event) error {

	// Convert event into discord payload
	payload := RenderDiscordEmbed(e)

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord marshal: %w", err)
	}

	// Create HTTP POST request to Discord webhook URL
	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook failed: status=%s", resp.Status)
	}

	return nil
}

//ensures string do not exceed Discord's maximum allowed field length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	if max <= 3 {
		return s[:max]
	}

	return s[:max-3] + "..."
}