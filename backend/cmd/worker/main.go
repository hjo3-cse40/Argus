package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"argus-backend/internal/config"
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

		// Parse JSON to extract event_id
		var payload map[string]any
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			log.Printf("json unmarshal error: %v", err)
			// If it's not valid JSON, Ack anyway to avoid getting stuck
			_ = msg.Ack(false)
			continue
		}

		eventID, _ := payload["event_id"].(string)
		log.Printf("RECEIVED event_id=%s", eventID)

		// Mark delivered back in API (best-effort)
		if eventID != "" {
			body, _ := json.Marshal(map[string]string{"event_id": eventID})
			req, _ := http.NewRequest("POST", apiBase+"/debug/delivered", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("mark delivered request error: %v", err)
			} else {
				_ = resp.Body.Close()
				log.Printf("marked delivered in API: status=%s", resp.Status)
			}
		}

		// Dummy delivery complete -> Ack message
		if err := msg.Ack(false); err != nil {
			log.Printf("ack error: %v", err)
		} else {
			log.Println("DELIVERED + ACKED")
		}
	}
}

