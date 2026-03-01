package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

type MarkFailedHandler struct {
	Store store.Store
}

func NewMarkFailedHandler(st store.Store) *MarkFailedHandler {
	return &MarkFailedHandler{Store: st}
}

func (h *MarkFailedHandler) Mark(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EventID     string `json:"event_id"`
		RetryCount  int    `json:"retry_count"`
		Error       string `json:"error"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EventID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "invalid body"})
		return
	}

	found := h.Store.MarkFailed(body.EventID, body.RetryCount, body.Error)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "found": found})
}