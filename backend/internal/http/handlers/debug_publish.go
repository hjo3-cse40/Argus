package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

type DebugPublisher struct {
	MQ    *mq.Client
	Store *store.MemoryStore
}

func NewDebugPublisher(mqClient *mq.Client, st *store.MemoryStore) *DebugPublisher {
	return &DebugPublisher{MQ: mqClient, Store: st}
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *DebugPublisher) Publish(w http.ResponseWriter, r *http.Request) {
	eventID := randomID()

	d := store.Delivery{
		EventID: eventID,
		Source:  "synthetic",
		Title:   "hello from argus",
		URL:     "https://example.com",
	}

	// Save as queued (in-memory)
	h.Store.AddQueued(d)

	// Publish to RabbitMQ
	event := map[string]any{
		"event_id":   eventID,
		"source":     d.Source,
		"title":      d.Title,
		"url":        d.URL,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(event)

	if err := h.MQ.Publish("raw_events", body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"event_id": eventID,
	})
}

