package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/notifier"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting worker in %s environment", cfg.Environment)

	amqpURL := cfg.RabbitMQ.URL
	apiBase := cfg.API.BaseURL

	discordWebhook := cfg.Destinations.DiscordWebhookURL
	if discordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK_URL not set")
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("channel:", err)
	}
	defer ch.Close()

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

	msgs, err := ch.Consume(
		queue,
		"",
		false, // autoAck = false -> we will Ack manually
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

	for msg := range msgs {
		log.Printf("RECEIVED raw message: %s", string(msg.Body))

		// parse event
		ev, err := events.FromJSON(msg.Body)
		if err != nil {
			log.Printf("event parse error: %v", err)
			_ = msg.Ack(false)
			continue
		}

		// Validate event
		if err := ev.Validate(); err != nil {
			log.Printf("event validation error: %v", err)
			_ = msg.Ack(false)
			continue
		}

		log.Printf("RECEIVED event_id=%s", ev.EventID)

		// Send to Discord
		if err := notifier.SendDiscordWebhook(discordWebhook, ev); err != nil {
			log.Printf("discord send failed event_id=%s err=%v", ev.EventID, err)
			_ = msg.Ack(false)
			continue
		}
		
		log.Printf("discord delivered event_id=%s", ev.EventID)

		// Mark delivered back in API (best-effort)
		body, _ := json.Marshal(map[string]string{"event_id": ev.EventID})
		req, _ := http.NewRequest("POST", apiBase+"/debug/delivered", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("mark delivered request error: %v", err)
		} else {
			_ = resp.Body.Close()
			log.Printf("marked delivered in API: status=%s", resp.Status)
		}


		// Dummy delivery complete -> Ack message
		if err := msg.Ack(false); err != nil {
			log.Printf("ack error: %v", err)
		} else {
			log.Println("DELIVERED + ACKED")
		}
	}
}

