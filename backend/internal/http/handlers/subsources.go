package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"argus-backend/internal/store"
)

type SubsourcesHandler struct {
	Store store.Store
}

func NewSubsourcesHandler(st store.Store) *SubsourcesHandler {
	return &SubsourcesHandler{Store: st}
}

// Create handles POST /api/platforms/:platform_id/subsources
func (h *SubsourcesHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extract platform_id from path
	platformID := r.PathValue("platform_id")
	if platformID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform ID is required",
		})
		return
	}

	// Verify platform exists
	_, found := h.Store.GetPlatform(platformID)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid platform",
			Details: []string{"platform_id does not reference an existing platform"},
		})
		return
	}

	// Parse JSON request body
	var req CreateSubsourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid JSON",
			Details: []string{err.Error()},
		})
		return
	}

	// Validate request
	if valErr := req.Validate(); valErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Validation failed",
			Details: valErr.Details,
		})
		return
	}

	// Convert to store.Subsource
	subsource := req.toStoreSubsource(platformID)

	// Add to store
	if err := h.Store.AddSubsource(subsource); err != nil {
		// Check if it's a duplicate identifier error
		if strings.Contains(err.Error(), "already exists") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Subsource already exists",
				Details: []string{err.Error()},
			})
			return
		}

		// Check if it's a platform not found error
		if strings.Contains(err.Error(), "platform not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Invalid platform",
				Details: []string{"platform_id does not reference an existing platform"},
			})
			return
		}

		log.Printf("Failed to add subsource: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Retrieve the created subsource (with generated ID and timestamp)
	subsources := h.Store.ListSubsources(platformID)
	if len(subsources) == 0 {
		log.Printf("Subsource was added but not found in store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Get the most recently added subsource (first in list since ordered by created_at DESC)
	createdSubsource := subsources[0]

	// Convert to response
	response := toSubsourceResponse(createdSubsource)

	// Return 201 Created with the subsource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// ListByPlatform handles GET /api/platforms/:platform_id/subsources
func (h *SubsourcesHandler) ListByPlatform(w http.ResponseWriter, r *http.Request) {
	// Extract platform_id from path
	platformID := r.PathValue("platform_id")
	if platformID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform ID is required",
		})
		return
	}

	// Retrieve subsources for platform from store
	subsources := h.Store.ListSubsources(platformID)

	// Convert to response format
	responses := make([]SubsourceResponse, len(subsources))
	for i, subsource := range subsources {
		responses[i] = toSubsourceResponse(subsource)
	}

	// Return 200 OK with subsources array
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}

// Get handles GET /api/subsources/:id
func (h *SubsourcesHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Extract id from path
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource ID is required",
		})
		return
	}

	// Retrieve subsource from store
	subsource, found := h.Store.GetSubsource(id)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource not found",
		})
		return
	}

	// Convert to response
	response := toSubsourceResponse(subsource)

	// Return 200 OK with subsource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Update handles PUT /api/subsources/:id
func (h *SubsourcesHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract id from path
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource ID is required",
		})
		return
	}

	// Get existing subsource to preserve platform_id, identifier, and created_at
	existingSubsource, found := h.Store.GetSubsource(id)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource not found",
		})
		return
	}

	// Parse JSON request body
	var req UpdateSubsourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Invalid JSON",
			Details: []string{err.Error()},
		})
		return
	}

	// Validate request
	if valErr := req.Validate(); valErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Validation failed",
			Details: valErr.Details,
		})
		return
	}

	// Create subsource update (preserve platform_id and created_at, allow identifier changes)
	subsource := store.Subsource{
		PlatformID: existingSubsource.PlatformID,
		Name:       req.Name,
		Identifier: req.Identifier,
		URL:        req.URL,
	}

	// Update in store
	if err := h.Store.UpdateSubsource(id, subsource); err != nil {
		log.Printf("Failed to update subsource: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Retrieve the updated subsource
	updatedSubsource, found := h.Store.GetSubsource(id)
	if !found {
		log.Printf("Subsource was updated but not found in store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Convert to response
	response := toSubsourceResponse(updatedSubsource)

	// Return 200 OK with updated subsource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Delete handles DELETE /api/subsources/:id
func (h *SubsourcesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract id from path
	id := r.PathValue("id")
	log.Printf("DELETE /api/subsources/%s - Request received", id)
	
	if id == "" {
		log.Printf("DELETE /api/subsources - Missing subsource ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource ID is required",
		})
		return
	}

	// Check if subsource exists
	_, found := h.Store.GetSubsource(id)
	if !found {
		log.Printf("DELETE /api/subsources/%s - Subsource not found", id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Subsource not found",
		})
		return
	}

	log.Printf("DELETE /api/subsources/%s - Subsource found, attempting delete", id)
	
	// Delete from store
	if err := h.Store.DeleteSubsource(id); err != nil {
		log.Printf("DELETE /api/subsources/%s - Failed to delete: %v", id, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	log.Printf("DELETE /api/subsources/%s - Successfully deleted", id)
	
	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
