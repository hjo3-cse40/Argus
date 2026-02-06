package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

type MarkDeliveredHandler struct {
	Store *store.MemoryStore
}

func NewMarkDeliveredHandler(st *store.MemoryStore) *MarkDeliveredHandler {
	return &MarkDeliveredHandler{Store: st}
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
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "found": found})
}
