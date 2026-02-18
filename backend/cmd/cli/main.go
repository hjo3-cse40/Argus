package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/mq"
)

func main() {
	var (
		source  = flag.String("source", "cli", "Event source identifier")
		title   = flag.String("title", "Test event from CLI", "Event title")
		url     = flag.String("url", "https://example.com", "Event URL")
		eventID = flag.String("event-id", "", "Custom event ID (generated if not provided)")
		count   = flag.Int("count", 1, "Number of events to publish")
		apiURL  = flag.String("api-url", "", "API base URL (overrides config)")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Override API URL if provided
	apiBaseURL := cfg.API.BaseURL
	if *apiURL != "" {
		apiBaseURL = *apiURL
	}

	// Connect to RabbitMQ
	mqClient, err := mq.Connect(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	// Ensure queue exists
	if err := mqClient.DeclareQueue("raw_events"); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < *count; i++ {
		// Generate event ID if not provided
		id := *eventID
		if id == "" || i > 0 {
			id = generateEventID()
		}

		// Create event
		event := events.NewEvent(id, *source, *title, *url)
		if err := event.Validate(); err != nil {
			log.Fatalf("invalid event: %v", err)
		}

		// Convert to JSON
		body, err := event.ToJSON()
		if err != nil {
			log.Fatalf("failed to marshal event: %v", err)
		}

		// Publish to RabbitMQ
		if err := mqClient.Publish("raw_events", body); err != nil {
			log.Fatalf("failed to publish event: %v", err)
		}

		fmt.Printf("✓ Published event: event_id=%s source=%s title=%s\n", id, *source, *title)

		// Optionally mark as queued in API (best-effort)
		if apiBaseURL != "" {
			markQueued(client, apiBaseURL, event)
		}

		// Small delay between events if publishing multiple
		if *count > 1 && i < *count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Printf("\n✓ Successfully published %d event(s)\n", *count)
}

func generateEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func markQueued(client *http.Client, apiBaseURL string, event *events.Event) {
	// This is a best-effort call to notify the API about the event
	// The API might have its own tracking mechanism
	payload := map[string]interface{}{
		"event_id": event.EventID,
		"source":   event.Source,
		"title":    event.Title,
		"url":      event.URL,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiBaseURL+"/debug/queued", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Ignore errors - this is best-effort
	_, _ = client.Do(req)
}
