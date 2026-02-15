package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"argus-backend/internal/store"
)

type SourcesHandler struct {
	Store *store.MemoryStore
}

func NewSourcesHandler(st *store.MemoryStore) *SourcesHandler {
	return &SourcesHandler{Store: st}
}

// Create handles POST /api/sources
func (h *SourcesHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var req CreateSourceRequest
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
	if err := req.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		// Extract details from ValidationError if available
		var details []string
		if valErr, ok := err.(*ValidationError); ok {
			details = valErr.Details
		} else {
			details = []string{err.Error()}
		}

		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Validation failed",
			Details: details,
		})
		return
	}

	// Convert to store.Source
	source := req.toStoreSource()

	// Add to store
	if err := h.Store.AddSource(source); err != nil {
		log.Printf("Failed to add source: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Retrieve the created source (with generated ID and timestamp)
	sources := h.Store.ListSources()
	if len(sources) == 0 {
		log.Printf("Source was added but not found in store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	// Get the most recently added source (last in list)
	createdSource := sources[len(sources)-1]

	// Convert to response (excludes webhook_secret)
	response := toSourceResponse(createdSource)

	// Return 201 Created with the source
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// List handles GET /api/sources
func (h *SourcesHandler) List(w http.ResponseWriter, r *http.Request) {
	// Check for name query parameter (for worker routing)
	name := r.URL.Query().Get("name")
	
	if name != "" {
		// Look up specific source by name
		source, found := h.Store.GetSourceByName(name)
		if !found {
			// Return empty array if not found
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]SourceResponse{})
			return
		}

		// Return single source in array
		response := toSourceResponse(source)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]SourceResponse{response})
		return
	}

	// Retrieve all sources from store
	sources := h.Store.ListSources()

	// Convert to response format (excludes webhook_secret)
	responses := make([]SourceResponse, len(sources))
	for i, source := range sources {
		responses[i] = toSourceResponse(source)
	}

	// Return 200 OK with sources array
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses)
}
