package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"argus-backend/internal/store"
)

type deliveryTestStore struct {
	mu         sync.Mutex
	deliveries []store.Delivery
}

func (s *deliveryTestStore) AddQueued(d store.Delivery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	d.Status = store.StatusQueued
	d.CreatedAt = now
	d.UpdatedAt = now
	s.deliveries = append([]store.Delivery{d}, s.deliveries...)
}

func (s *deliveryTestStore) MarkDelivered(eventID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.deliveries {
		if s.deliveries[i].EventID == eventID {
			s.deliveries[i].Status = store.StatusDelivered
			s.deliveries[i].UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

func (s *deliveryTestStore) MarkFailed(eventID string, retryCount int, lastErr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.deliveries {
		if s.deliveries[i].EventID == eventID {
			s.deliveries[i].Status = store.StatusFailed
			s.deliveries[i].RetryCount = retryCount
			s.deliveries[i].LastError = lastErr
			s.deliveries[i].UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

func (s *deliveryTestStore) List() []store.Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]store.Delivery, len(s.deliveries))
	copy(out, s.deliveries)
	return out
}

func (s *deliveryTestStore) ListDeliveriesBySubsource(subsourceID string) []store.Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []store.Delivery
	for _, d := range s.deliveries {
		if d.SubsourceID != nil && *d.SubsourceID == subsourceID {
			out = append(out, d)
		}
	}
	return out
}

func (s *deliveryTestStore) ListDeliveriesByPlatform(string) []store.Delivery { return nil }

func (s *deliveryTestStore) AddSource(store.Source) error                          { return nil }
func (s *deliveryTestStore) ListSources() []store.Source                           { return nil }
func (s *deliveryTestStore) GetSource(string) (store.Source, bool)                 { return store.Source{}, false }
func (s *deliveryTestStore) GetSourceByName(string) (store.Source, bool)           { return store.Source{}, false }
func (s *deliveryTestStore) AddPlatform(store.Platform) error                      { return nil }
func (s *deliveryTestStore) ListPlatforms() []store.Platform                       { return nil }
func (s *deliveryTestStore) GetPlatform(string) (store.Platform, bool)             { return store.Platform{}, false }
func (s *deliveryTestStore) GetPlatformByName(string) (store.Platform, bool)       { return store.Platform{}, false }
func (s *deliveryTestStore) UpdatePlatform(string, store.Platform) error           { return nil }
func (s *deliveryTestStore) DeletePlatform(string) error                           { return nil }
func (s *deliveryTestStore) AddSubsource(store.Subsource) error                    { return nil }
func (s *deliveryTestStore) ListSubsources(string) []store.SubsourceWithPlatform   { return nil }
func (s *deliveryTestStore) ListAllSubsources() []store.SubsourceWithPlatform      { return nil }
func (s *deliveryTestStore) GetSubsource(string) (store.SubsourceWithPlatform, bool) {
	return store.SubsourceWithPlatform{}, false
}
func (s *deliveryTestStore) UpdateSubsource(string, store.Subsource) error { return nil }
func (s *deliveryTestStore) DeleteSubsource(string) error                  { return nil }
func (s *deliveryTestStore) AddFilter(store.DestinationFilter) error       { return nil }
func (s *deliveryTestStore) ListFilters(string) []store.DestinationFilter  { return nil }
func (s *deliveryTestStore) DeleteFilter(string) error                     { return nil }
func (s *deliveryTestStore) CreateUser(store.User) error                   { return nil }
func (s *deliveryTestStore) GetUserByEmail(string) (store.User, bool)      { return store.User{}, false }
func (s *deliveryTestStore) GetUserByID(string) (store.User, bool)         { return store.User{}, false }
func (s *deliveryTestStore) CreateSession(store.Session) error             { return nil }
func (s *deliveryTestStore) GetSession(string) (store.Session, bool)       { return store.Session{}, false }
func (s *deliveryTestStore) DeleteSession(string) error                    { return nil }
func (s *deliveryTestStore) DeleteExpiredSessions() error                  { return nil }
func (s *deliveryTestStore) Close() error                                  { return nil }

func newTestStore() *deliveryTestStore {
	return &deliveryTestStore{}
}

func seedDeliveries(st *deliveryTestStore) {
	sub := "sub-1"
	entries := []struct {
		eventID string
		source  string
		title   string
		markAs  string
	}{
		{"evt-1", "youtube", "New Video: Go Testing", "delivered"},
		{"evt-2", "reddit", "Post in r/golang", "delivered"},
		{"evt-3", "x", "Tweet from @gopher", "failed"},
		{"evt-4", "youtube", "New Video: Concurrency", "queued"},
		{"evt-5", "reddit", "Post in r/programming", "delivered"},
	}
	for _, e := range entries {
		st.AddQueued(store.Delivery{
			EventID:     e.eventID,
			Source:      e.source,
			Title:       e.title,
			URL:         "https://example.com/" + e.eventID,
			SubsourceID: &sub,
		})
		switch e.markAs {
		case "delivered":
			st.MarkDelivered(e.eventID)
		case "failed":
			st.MarkFailed(e.eventID, 3, "webhook timeout")
		}
		time.Sleep(time.Millisecond)
	}
}

// US 3.7: Empty history returns empty JSON array (UI shows "No notifications yet.")
func TestDeliveriesHandler_List_Empty(t *testing.T) {
	st := newTestStore()
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) != 0 {
		t.Errorf("Expected 0 deliveries, got %d", len(deliveries))
	}
}

// US 3.7: All deliveries appear in notification history
func TestDeliveriesHandler_List_ReturnsAll(t *testing.T) {
	st := newTestStore()
	seedDeliveries(st)
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) != 5 {
		t.Errorf("Expected 5 deliveries, got %d", len(deliveries))
	}
}

// US 3.7: JSON response contains all fields the UI needs to render the table
func TestDeliveriesHandler_List_FieldsForUI(t *testing.T) {
	st := newTestStore()
	st.AddQueued(store.Delivery{
		EventID: "evt-ui",
		Source:  "youtube",
		Title:   "UI Field Test",
		URL:     "https://example.com/ui",
	})
	st.MarkDelivered("evt-ui")
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("Expected 1 delivery, got %d", len(deliveries))
	}

	d := deliveries[0]
	if d.Source != "youtube" {
		t.Errorf("Expected source 'youtube', got %q", d.Source)
	}
	if d.Status != store.StatusDelivered {
		t.Errorf("Expected status 'delivered', got %q", d.Status)
	}
	if d.CreatedAt.IsZero() {
		t.Error("Expected non-zero created_at")
	}
	if d.UpdatedAt.IsZero() {
		t.Error("Expected non-zero updated_at")
	}
}

// US 3.8: Status filter dropdown works for each status
func TestDeliveriesHandler_List_FilterByStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected int
	}{
		{"delivered only", "delivered", 3},
		{"failed only", "failed", 1},
		{"queued only", "queued", 1},
		{"no filter returns all", "", 5},
		{"invalid filter returns none", "unknown", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := newTestStore()
			seedDeliveries(st)
			handler := NewDeliveriesHandler(st)

			url := "/deliveries"
			if tc.status != "" {
				url += "?status=" + tc.status
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected 200, got %d", w.Code)
			}

			var deliveries []store.Delivery
			if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}
			if len(deliveries) != tc.expected {
				t.Errorf("Expected %d deliveries for status=%q, got %d", tc.expected, tc.status, len(deliveries))
			}
		})
	}
}

// US 3.7: Default limit is 50, explicit limit truncates
func TestDeliveriesHandler_List_Limit(t *testing.T) {
	st := newTestStore()
	seedDeliveries(st)
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries?limit=2", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) != 2 {
		t.Errorf("Expected 2 deliveries with limit=2, got %d", len(deliveries))
	}
}

// Limit is capped at maxDeliveryLimit (100) even if client asks for more
func TestDeliveriesHandler_List_LimitCappedAt100(t *testing.T) {
	st := newTestStore()
	for i := 0; i < 150; i++ {
		st.AddQueued(store.Delivery{
			EventID: "evt-cap-" + string(rune(i+33)),
			Source:  "test",
			Title:   "Cap Test",
		})
	}
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries?limit=150", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) > 100 {
		t.Errorf("Expected max 100 deliveries, got %d", len(deliveries))
	}
}

// All returned statuses are valid — the UI maps these to Success/Failed/Pending labels
func TestDeliveriesHandler_List_StatusLabelsMatchUI(t *testing.T) {
	st := newTestStore()
	seedDeliveries(st)
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	validStatuses := map[store.DeliveryStatus]bool{
		store.StatusDelivered: true,
		store.StatusFailed:    true,
		store.StatusQueued:    true,
	}

	for _, d := range deliveries {
		if !validStatuses[d.Status] {
			t.Errorf("Unexpected status %q for event %s — UI won't render it correctly", d.Status, d.EventID)
		}
	}
}

// JSON keys must match what the frontend reads: source, status, created_at, updated_at
func TestDeliveriesHandler_List_JSONShape(t *testing.T) {
	st := newTestStore()
	st.AddQueued(store.Delivery{
		EventID: "evt-shape",
		Source:  "reddit",
		Title:   "Shape Test",
		URL:     "https://reddit.com/r/test",
	})
	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var raw []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(raw) != 1 {
		t.Fatalf("Expected 1 delivery, got %d", len(raw))
	}

	requiredFields := []string{"event_id", "source", "status", "created_at", "updated_at"}
	for _, field := range requiredFields {
		if _, ok := raw[0][field]; !ok {
			t.Errorf("Missing required JSON field %q — UI notification history depends on it", field)
		}
	}
}

// Filter by subsource_id — used when drilling into a specific subsource
func TestDeliveriesHandler_List_FilterBySubsource(t *testing.T) {
	st := newTestStore()
	sub1 := "sub-1"
	sub2 := "sub-2"

	st.AddQueued(store.Delivery{EventID: "e1", Source: "youtube", SubsourceID: &sub1})
	st.AddQueued(store.Delivery{EventID: "e2", Source: "reddit", SubsourceID: &sub2})
	st.AddQueued(store.Delivery{EventID: "e3", Source: "youtube", SubsourceID: &sub1})

	handler := NewDeliveriesHandler(st)

	req := httptest.NewRequest("GET", "/deliveries?subsource_id=sub-1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var deliveries []store.Delivery
	if err := json.NewDecoder(w.Body).Decode(&deliveries); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(deliveries) != 2 {
		t.Errorf("Expected 2 deliveries for sub-1, got %d", len(deliveries))
	}
}
