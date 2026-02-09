package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"argus-backend/internal/events"
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

	// Create event using the formalized schema
	event := events.NewEvent(eventID, "synthetic", "hello from argus", "https://example.com")

	// Validate event
	if err := event.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Convert to store.Delivery for in-memory tracking
	d := store.Delivery{
		EventID: event.EventID,
		Source:  event.Source,
		Title:   event.Title,
		URL:     event.URL,
	}

	// Save as queued (in-memory)
	h.Store.AddQueued(d)

	// Convert event to JSON and publish to RabbitMQ
	body, err := event.ToJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "failed to marshal event",
		})
		return
	}

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

