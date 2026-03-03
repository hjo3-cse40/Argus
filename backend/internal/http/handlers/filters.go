package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"argus-backend/internal/store"
)

type FiltersHandler struct {
	Store store.Store
}

func NewFiltersHandler(st store.Store) *FiltersHandler {
	return &FiltersHandler{Store: st}
}

type CreateFilterRequest struct {
	FilterType string `json:"filter_type"`
	Pattern    string `json:"pattern"`
}

type FilterResponse struct {
	ID         string    `json:"id"`
	PlatformID string    `json:"platform_id"`
	FilterType string    `json:"filter_type"`
	Pattern    string    `json:"pattern"`
	CreatedAt  time.Time `json:"created_at"`
}

func (r *CreateFilterRequest) Validate() *ValidationError {
	var details []string

	allowed := map[string]bool{
		"keyword_include": true,
		"keyword_exclude": true,
	}
	if strings.TrimSpace(r.FilterType) == "" {
		details = append(details, "filter_type is required")
	} else if !allowed[r.FilterType] {
		details = append(details, "filter_type must be one of: keyword_include, keyword_exclude")
	}

	if strings.TrimSpace(r.Pattern) == "" {
		details = append(details, "pattern is required")
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}
	return nil
}

// Create handles POST /api/platforms/{platform_id}/filters
func (h *FiltersHandler) Create(w http.ResponseWriter, r *http.Request) {
	platformID := r.PathValue("platform_id")
	if platformID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "platform_id is required"})
		return
	}

	_, found := h.Store.GetPlatform(platformID)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Platform not found"})
		return
	}

	var req CreateFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid JSON",
			Details: []string{err.Error()},
		})
		return
	}

	if valErr := req.Validate(); valErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Validation failed",
			Details: valErr.Details,
		})
		return
	}

	f := store.DestinationFilter{
		PlatformID: platformID,
		FilterType: req.FilterType,
		Pattern:    req.Pattern,
	}

	if err := h.Store.AddFilter(f); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create filter"})
		return
	}

	// Return the most recently added filter for this platform
	filters := h.Store.ListFilters(platformID)
	if len(filters) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
		return
	}

	created := filters[0]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toFilterResponse(created))
}

// List handles GET /api/platforms/{platform_id}/filters
func (h *FiltersHandler) List(w http.ResponseWriter, r *http.Request) {
	platformID := r.PathValue("platform_id")
	if platformID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "platform_id is required"})
		return
	}

	filters := h.Store.ListFilters(platformID)

	responses := make([]FilterResponse, len(filters))
	for i, f := range filters {
		responses[i] = toFilterResponse(f)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}

// Delete handles DELETE /api/filters/{id}
func (h *FiltersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Filter ID is required"})
		return
	}

	if err := h.Store.DeleteFilter(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete filter"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toFilterResponse(f store.DestinationFilter) FilterResponse {
	return FilterResponse{
		ID:         f.ID,
		PlatformID: f.PlatformID,
		FilterType: f.FilterType,
		Pattern:    f.Pattern,
		CreatedAt:  f.CreatedAt,
	}
}
