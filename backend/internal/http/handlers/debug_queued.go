package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

// MarkQueuedHandler records an event as queued in the store only (no MQ publish).
// Used by the CLI after it has already published to RabbitMQ.
type MarkQueuedHandler struct {
	Store store.Store
}

func NewMarkQueuedHandler(st store.Store) *MarkQueuedHandler {
	return &MarkQueuedHandler{Store: st}
}

func (h *MarkQueuedHandler) MarkQueued(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EventID string `json:"event_id"`
		Source  string `json:"source"`
		Title   string `json:"title"`
		URL     string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EventID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "invalid body: need event_id"})
		return
	}

	d := store.Delivery{
		EventID: body.EventID,
		Source:  body.Source,
		Title:   body.Title,
		URL:     body.URL,
	}
	h.Store.AddQueued(d)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "event_id": body.EventID})
}
