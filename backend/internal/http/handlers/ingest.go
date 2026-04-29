package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"argus-backend/internal/events"
	"argus-backend/internal/filter"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

// IngestHandler accepts normalized event payloads, validates them, publishes to RabbitMQ, and records as queued.
type IngestHandler struct {
	MQ    *mq.Client
	Store store.Store
}

// NewIngestHandler returns a new IngestHandler.
func NewIngestHandler(mqClient *mq.Client, st store.Store) *IngestHandler {
	return &IngestHandler{MQ: mqClient, Store: st}
}

func randomEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Ingest handles POST /api/ingest. Body must be JSON in Argus event shape.
// event_id and created_at are optional (generated/set if missing).
func (h *IngestHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var ev events.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "invalid JSON body",
		})
		return
	}

	// Normalize: fill event_id and created_at if missing
	if ev.EventID == "" {
		ev.EventID = randomEventID()
	}
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now().UTC()
	}
	if ev.Metadata == nil {
		ev.Metadata = make(map[string]interface{})
	}

	if err := ev.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	body, err := ev.ToJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "failed to marshal event",
		})
		return
	}

	if err := h.MQ.Publish("raw_events", body); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "queue unavailable",
		})
		return
	}

	// Extract subsource_id from event metadata if present
	var subsourceID *string
	if sid, ok := ev.Metadata["subsource_id"].(string); ok && sid != "" {
		subsourceID = &sid
	}

	// If the event is tied to a known subsource/platform, evaluate keyword filters
	// before enqueueing so blocked events can be reported immediately to the caller.
	if subsourceID != nil {
		if subsource, found := h.Store.GetSubsource(*subsourceID); found {
			filters := h.Store.ListFilters(subsource.PlatformID)
			if pass, reason := filter.EvaluateWithReason(&ev, filters); !pass {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"ok":       false,
					"blocked":  true,
					"event_id": ev.EventID,
					"reason":   reason,
				})
				return
			}
		}
	}

	h.Store.AddQueued(store.Delivery{
		EventID:     ev.EventID,
		Source:      ev.Source,
		Title:       ev.Title,
		URL:         ev.URL,
		SubsourceID: subsourceID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"event_id": ev.EventID,
	})
}
