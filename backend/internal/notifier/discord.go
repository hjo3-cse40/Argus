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

func RenderDiscordEmbed(e *events.Event) discordPayload {
	sourceType, _ := e.Metadata["source_type"].(string)
	author, _ := e.Metadata["author"].(string)
	description, _ := e.Metadata["description"].(string)

	var descParts []string
	if sourceType != "" {
		descParts = append(descParts, fmt.Sprintf("**Platform:** %s", sourceType))
	}
	descParts = append(descParts, fmt.Sprintf("**Source:** %s", e.Source))
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