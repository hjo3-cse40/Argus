package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

type MarkDeliveredHandler struct {
	Store       store.Store
	Broadcaster *DeliveryBroadcaster
}

func NewMarkDeliveredHandler(st store.Store, broadcaster *DeliveryBroadcaster) *MarkDeliveredHandler {
	return &MarkDeliveredHandler{Store: st, Broadcaster: broadcaster}
}

func (h *MarkDeliveredHandler) Mark(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EventID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "invalid body"})
		return
	}

	found := h.Store.MarkDelivered(body.EventID)

	if found && h.Broadcaster != nil {
		for _, d := range h.Store.List() {
			if d.EventID == body.EventID && d.Status == store.StatusDelivered {
				h.Broadcaster.Publish(d)
				break
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "found": found})
}
