package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

type DeliveriesHandler struct {
	Store *store.MemoryStore
}

func NewDeliveriesHandler(st *store.MemoryStore) *DeliveriesHandler {
	return &DeliveriesHandler{Store: st}
}

func (h *DeliveriesHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.Store.List())
}
