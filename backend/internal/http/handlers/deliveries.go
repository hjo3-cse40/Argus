package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"argus-backend/internal/auth"
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

// sortDeliveriesInPlace stable-sorts list by sort (created_at, updated_at, title, source)
// and order (asc, desc). Invalid sort defaults to created_at. Missing/invalid order uses
// desc for times and asc for title/source.
func sortDeliveriesInPlace(list []store.Delivery, sortField, order string) {
	if len(list) < 2 {
		return
	}
	field := strings.ToLower(strings.TrimSpace(sortField))
	switch field {
	case "created_at", "updated_at", "title", "source":
	default:
		field = "created_at"
	}
	ord := strings.ToLower(strings.TrimSpace(order))
	var asc bool
	switch ord {
	case "asc":
		asc = true
	case "desc":
		asc = false
	default:
		asc = field == "title" || field == "source"
	}

	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		switch field {
		case "updated_at":
			if asc {
				return a.UpdatedAt.Before(b.UpdatedAt)
			}
			return a.UpdatedAt.After(b.UpdatedAt)
		case "title":
			ca, cb := strings.ToLower(a.Title), strings.ToLower(b.Title)
			if asc {
				return ca < cb
			}
			return ca > cb
		case "source":
			ca, cb := strings.ToLower(a.Source), strings.ToLower(b.Source)
			if asc {
				return ca < cb
			}
			return ca > cb
		default: // created_at
			if asc {
				return a.CreatedAt.Before(b.CreatedAt)
			}
			return a.CreatedAt.After(b.CreatedAt)
		}
	})
}

func (h *DeliveriesHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
		deliveries = h.Store.ListDeliveriesBySubsource(user.ID, subsourceID)
	} else if platformID != "" {
		deliveries = h.Store.ListDeliveriesByPlatform(user.ID, platformID)
	} else {
		deliveries = h.Store.List(user.ID)
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

	sortField := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")
	sortDeliveriesInPlace(deliveries, sortField, sortOrder)

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
