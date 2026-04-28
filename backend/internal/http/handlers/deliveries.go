package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"argus-backend/internal/store"
)

//Default limit of last 50 notifications
const defaultDeliveryLimit = 50
const maxDeliveryLimit = 100

type DeliveriesHandler struct {
	Store store.Store
}

func NewDeliveriesHandler(st store.Store) *DeliveriesHandler {
	return &DeliveriesHandler{Store: st}
}

func (h *DeliveriesHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	
	subsourceID := r.URL.Query().Get("subsource_id")
	platformID := r.URL.Query().Get("platform_id")
	// Check for filtering query parameters(status: success, pending, failed)
	statusFilter := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status"))) // US 3.8

	limit := defaultDeliveryLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			if n > maxDeliveryLimit {
				n = maxDeliveryLimit
			}
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	//Fetch deliveries from store
	var deliveries []store.Delivery

	if subsourceID != "" {
		deliveries = h.Store.ListDeliveriesBySubsource(subsourceID)
	} else if platformID != "" {
		deliveries = h.Store.ListDeliveriesByPlatform(platformID)
	} else {
		deliveries = h.Store.List()
	}

	// Filter by status if requested
	if statusFilter != "" {
		if statusFilter == "delivered" || statusFilter == "failed" || statusFilter == "queued" {
			filtered := make([]store.Delivery, 0, len(deliveries))
			for _, d := range deliveries {
				if strings.EqualFold(string(d.Status), statusFilter) {
					filtered = append(filtered, d)
				}
			}
			deliveries = filtered
		} else {
			deliveries = make([]store.Delivery, 0)
		}
	}

	// Apply offset + limit when list is ordered by created_at DESC
	if offset >= len(deliveries) {
		deliveries = make([]store.Delivery, 0)
	} else {
		deliveries = deliveries[offset:]
		if len(deliveries) > limit {
			deliveries = deliveries[:limit]
		}
	}
	//Return JSON
	_ = json.NewEncoder(w).Encode(deliveries)
}
