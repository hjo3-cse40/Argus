package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"

	"argus-backend/internal/auth"
	"argus-backend/internal/store"
)

type PlatformsHandler struct {
	Store store.Store
}

func NewPlatformsHandler(st store.Store) *PlatformsHandler {
	return &PlatformsHandler{Store: st}
}

// Create handles POST /api/platforms
func (h *PlatformsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse JSON request body
	var req CreatePlatformRequest
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

	// Convert to store.Platform
	platform := req.toStorePlatform()

	// Add to store
	if err := h.Store.AddPlatform(user.ID, platform); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Platform already exists",
				Details: []string{"A platform with this name already exists (names must be unique)."},
			})
			return
		}
		// Check if it's a duplicate name error
		if strings.Contains(err.Error(), "already exists") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Platform already exists",
				Details: []string{err.Error()},
			})
			return
		}

		log.Printf("Failed to add platform: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	createdPlatform, found := h.Store.GetPlatformByName(user.ID, platform.Name)
	if !found {
		log.Printf("Platform was added but not found in store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Convert to response (excludes webhook_secret)
	response := toPlatformResponse(createdPlatform)

	// Return 201 Created with the platform
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// List handles GET /api/platforms
func (h *PlatformsHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	platforms := h.Store.ListPlatforms(user.ID)

	// Convert to response format (excludes webhook_secret)
	responses := make([]PlatformResponse, len(platforms))
	for i, platform := range platforms {
		responses[i] = toPlatformResponse(platform)
	}

	// Return 200 OK with platforms array
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}

// Get handles GET /api/platforms/:id
func (h *PlatformsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract id from path
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform ID is required",
		})
		return
	}

	platform, found := h.Store.GetPlatform(user.ID, id)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform not found",
		})
		return
	}

	// Convert to response (excludes webhook_secret)
	response := toPlatformResponse(platform)

	// Return 200 OK with platform
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Update handles PUT /api/platforms/:id
func (h *PlatformsHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract id from path
	id := r.PathValue("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform ID is required",
		})
		return
	}

	existingPlatform, found := h.Store.GetPlatform(user.ID, id)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform not found",
		})
		return
	}

	// Parse JSON request body
	var req UpdatePlatformRequest
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

	includeCombine := existingPlatform.FilterIncludeCombine
	excludeCombine := existingPlatform.FilterExcludeCombine
	if req.FilterIncludeCombine != nil {
		includeCombine = *req.FilterIncludeCombine
	}
	if req.FilterExcludeCombine != nil {
		excludeCombine = *req.FilterExcludeCombine
	}

	// Create platform update (preserve name and created_at)
	platform := store.Platform{
		Name:                   existingPlatform.Name,
		DiscordWebhook:         req.DiscordWebhook,
		WebhookSecret:          req.WebhookSecret,
		FilterIncludeCombine:   includeCombine,
		FilterExcludeCombine:   excludeCombine,
	}

	if err := h.Store.UpdatePlatform(user.ID, id, platform); err != nil {
		log.Printf("Failed to update platform: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	updatedPlatform, found := h.Store.GetPlatform(user.ID, id)
	if !found {
		log.Printf("Platform was updated but not found in store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Convert to response (excludes webhook_secret)
	response := toPlatformResponse(updatedPlatform)

	// Return 200 OK with updated platform
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Delete handles DELETE /api/platforms/:id
func (h *PlatformsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	log.Printf("DELETE /api/platforms/%s - Request received", id)
	
	if id == "" {
		log.Printf("DELETE /api/platforms - Missing platform ID")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform ID is required",
		})
		return
	}

	_, found := h.Store.GetPlatform(user.ID, id)
	if !found {
		log.Printf("DELETE /api/platforms/%s - Platform not found", id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Platform not found",
		})
		return
	}

	log.Printf("DELETE /api/platforms/%s - Platform found, attempting delete", id)
	
	if err := h.Store.DeletePlatform(user.ID, id); err != nil {
		log.Printf("DELETE /api/platforms/%s - Failed to delete: %v", id, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	log.Printf("DELETE /api/platforms/%s - Successfully deleted", id)
	
	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
