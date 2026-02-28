package handlers

import (
	"encoding/json"
	"net/http"

	"argus-backend/internal/store"
)

type DeliveriesHandler struct {
	Store store.Store
}

func NewDeliveriesHandler(st store.Store) *DeliveriesHandler {
	return &DeliveriesHandler{Store: st}
}

func (h *DeliveriesHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check for filtering query parameters
	subsourceID := r.URL.Query().Get("subsource_id")
	platformID := r.URL.Query().Get("platform_id")
	
	var deliveries []store.Delivery
	
	if subsourceID != "" {
		// Filter by subsource_id
		deliveries = h.Store.ListDeliveriesBySubsource(subsourceID)
	} else if platformID != "" {
		// Filter by platform_id
		deliveries = h.Store.ListDeliveriesByPlatform(platformID)
	} else {
		// No filter, return all
		deliveries = h.Store.List()
	}
	
	_ = json.NewEncoder(w).Encode(deliveries)
}
