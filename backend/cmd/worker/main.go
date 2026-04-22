package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/filter"
	"argus-backend/internal/notifier"
	"argus-backend/internal/store"
	amqp "github.com/rabbitmq/amqp091-go"
)

// notifierStoreAdapter wraps store.Store to satisfy notifier.Store,
// converting between the store and notifier type systems.
type notifierStoreAdapter struct {
	st store.Store
}

func (a *notifierStoreAdapter) GetSubsource(id string) (notifier.Subsource, bool) {
	sub, found := a.st.GetSubsource(id)
	if !found {
		return notifier.Subsource{}, false
	}
	return notifier.Subsource{ID: sub.ID, PlatformID: sub.PlatformID, Name: sub.Name}, true
}

func (a *notifierStoreAdapter) GetPlatform(id string) (notifier.Platform, bool) {
	p, found := a.st.GetPlatform(id)
	if !found {
		return notifier.Platform{}, false
	}
	return notifier.Platform{ID: p.ID, Name: p.Name, DiscordWebhook: p.DiscordWebhook}, true
}

func (a *notifierStoreAdapter) GetSourceByName(name string) (notifier.Source, bool) {
	s, found := a.st.GetSourceByName(name)
	if !found {
		return notifier.Source{}, false
	}
	return notifier.Source{ID: s.ID, Name: s.Name, DiscordWebhook: s.DiscordWebhook}, true
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting worker in %s environment", cfg.Environment)

	amqpURL := cfg.RabbitMQ.URL
	apiBase := cfg.API.BaseURL

	fallbackWebhook := cfg.Destinations.DiscordWebhookURL
	if fallbackWebhook != "" {
		log.Printf("Fallback Discord webhook URL loaded")
	}

	// Connect to database for per-platform webhook resolution and filtering
	connStr := cfg.Database.ConnectionString()
	st, err := store.NewPostgresStore(connStr, cfg.DeliveryLimit)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() { _ = st.Close() }()

	noti := notifier.NewNotifier(&notifierStoreAdapter{st: st})

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("channel:", err)
	}
	defer func() { _ = ch.Close() }()

	//Declare the queue that we want to consume: raw_events
	queue := "raw_events"
	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("queue declare:", err)
	}
	//Consume the queue
	msgs, err := ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("consume:", err)
	}

	log.Println("worker listening on raw_events")

	client := &http.Client{Timeout: 3 * time.Second}

	const maxRetries = 3
	baseDelay := 1 * time.Second
	
	for msg := range msgs {
		log.Printf("RECEIVED raw message: %s", string(msg.Body))
		//Parse the event
		ev, err := events.FromJSON(msg.Body)
		if err != nil {
			log.Printf("event parse error: %v", err)
			_ = msg.Ack(false)
			continue
		}

		if err := ev.Validate(); err != nil {
			log.Printf("event validation error: %v", err)
			_ = msg.Ack(false)
			continue
		}

		log.Printf("RECEIVED event_id=%s", ev.EventID)

		// Resolve per-platform webhook; fall back to global config
		webhookURL, platformID, resolveErr := noti.ResolveDestination(ev)
		if resolveErr != nil || webhookURL == "" {
			if fallbackWebhook != "" {
				log.Printf("destination resolve failed (%v), using fallback webhook", resolveErr)
				webhookURL = fallbackWebhook
				platformID = ""
			} else {
				log.Printf("destination resolve failed and no fallback: %v", resolveErr)
				_ = msg.Ack(false)
				continue
			}
		}

		// Apply per-destination filters
		if platformID != "" {
			filters := st.ListFilters(platformID)
			if !filter.Evaluate(ev, filters) {
				log.Printf("FILTERED event_id=%s platform_id=%s (did not pass destination filters)",
					ev.EventID, platformID)
				_ = msg.Ack(false)
				continue
			}
		}

		// Retry Discord delivery
		var lastErr error
		success := false
		attemptsMade := 0

		for attempt := 1; attempt <= maxRetries; attempt++ {
			attemptsMade = attempt

			err := notifier.SendDiscordWebhook(webhookURL, ev)
			if err == nil {
				success = true
				log.Printf("discord delivered event_id=%s attempt=%d/%d",
					ev.EventID, attempt, maxRetries)
				break
			}

			lastErr = err
			log.Printf("discord send failed event_id=%s attempt=%d/%d err=%v",
				ev.EventID, attempt, maxRetries, err)

			if attempt < maxRetries {
				time.Sleep(baseDelay * time.Duration(1<<(attempt-1)))
			}
		}

		// Record failure after retry limit
		if !success {
			errMsg := ""
			if lastErr != nil {
				errMsg = lastErr.Error()
			}

			failPayload := map[string]any{
				"event_id":    ev.EventID,
				"retry_count": attemptsMade,
				"error":       errMsg,
			}

			body, _ := json.Marshal(failPayload)
			req, _ := http.NewRequest("POST", apiBase+"/debug/failed", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("mark failed request error: %v", err)
			} else {
				_ = resp.Body.Close()
				log.Printf("marked failed in API: status=%s", resp.Status)
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("ack error: %v", err)
			} else {
				log.Println("FAILED + ACKED")
			}
			continue
		}

		// Mark delivered back in API
		deliveredBody, _ := json.Marshal(map[string]string{"event_id": ev.EventID})
		deliveredReq, _ := http.NewRequest("POST", apiBase+"/debug/delivered", bytes.NewReader(deliveredBody))
		deliveredReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(deliveredReq)
		if err != nil {
			log.Printf("mark delivered request error: %v", err)
		} else {
			_ = resp.Body.Close()
			log.Printf("marked delivered in API: status=%s", resp.Status)
		}

		if err := msg.Ack(false); err != nil {
			log.Printf("ack error: %v", err)
		} else {
			log.Println("DELIVERED + ACKED")
		}
	}
}
