package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"argus-backend/internal/store"
)

func TestSourcesHandler_Create_InvalidJSON(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/api/sources", strings.NewReader("{invalid json"))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error != "Invalid JSON" {
		t.Errorf("Expected error 'Invalid JSON', got %q", errResp.Error)
	}
}

func TestSourcesHandler_Create_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		payload     CreateSourceRequest
		wantErr     string
		wantDetails []string
	}{
		{
			name:        "missing name",
			payload:     CreateSourceRequest{Type: "github", DiscordWebhook: "https://discord.com/api/webhooks/123/abc"},
			wantErr:     "Validation failed",
			wantDetails: []string{"name is required"},
		},
		{
			name:        "missing type",
			payload:     CreateSourceRequest{Name: "Test", DiscordWebhook: "https://discord.com/api/webhooks/123/abc"},
			wantErr:     "Validation failed",
			wantDetails: []string{"type is required"},
		},
		{
			name:        "missing discord webhook",
			payload:     CreateSourceRequest{Name: "Test", Type: "github"},
			wantErr:     "Validation failed",
			wantDetails: []string{"discord_webhook is required"},
		},
		{
			name:        "all fields missing",
			payload:     CreateSourceRequest{},
			wantErr:     "Validation failed",
			wantDetails: []string{"name is required", "type is required", "discord_webhook is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := store.NewMemoryStore(100)
			handler := NewSourcesHandler(st)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/sources", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Create(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errResp.Error != tt.wantErr {
				t.Errorf("Expected error %q, got %q", tt.wantErr, errResp.Error)
			}

			for _, detail := range tt.wantDetails {
				found := false
				for _, d := range errResp.Details {
					if strings.Contains(d, detail) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected detail containing %q, got %v", detail, errResp.Details)
				}
			}
		})
	}
}

func TestSourcesHandler_Create_MalformedDiscordWebhook(t *testing.T) {
	tests := []struct {
		name    string
		webhook string
	}{
		{"http instead of https", "http://discord.com/api/webhooks/123/abc"},
		{"wrong domain", "https://example.com/api/webhooks/123/abc"},
		{"missing webhooks path", "https://discord.com/api/123/abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := store.NewMemoryStore(100)
			handler := NewSourcesHandler(st)

			payload := CreateSourceRequest{
				Name:           "Test Source",
				Type:           "generic",
				DiscordWebhook: tt.webhook,
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/sources", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Create(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errResp.Error != "Validation failed" {
				t.Errorf("Expected error 'Validation failed', got %q", errResp.Error)
			}

			found := false
			for _, detail := range errResp.Details {
				if strings.Contains(detail, "discord_webhook") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected detail about discord_webhook, got %v", errResp.Details)
			}
		})
	}
}

func TestSourcesHandler_Create_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	payload := CreateSourceRequest{
		Name:           "GitHub Main",
		Type:           "github",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
		WebhookSecret:  "secret123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/sources", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp SourceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if resp.Name != payload.Name {
		t.Errorf("Expected name %q, got %q", payload.Name, resp.Name)
	}
	if resp.Type != payload.Type {
		t.Errorf("Expected type %q, got %q", payload.Type, resp.Type)
	}
	if resp.DiscordWebhook != payload.DiscordWebhook {
		t.Errorf("Expected webhook %q, got %q", payload.DiscordWebhook, resp.DiscordWebhook)
	}
	if resp.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

func TestSourcesHandler_List_Empty(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	req := httptest.NewRequest("GET", "/api/sources", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []SourceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("Expected empty array, got %d items", len(resp))
	}
}

func TestSourcesHandler_List_WithSources(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	// Add some sources
	_ = st.AddSource(store.Source{
		Name:           "Source 1",
		Type:           "github",
		DiscordWebhook: "https://discord.com/api/webhooks/1/a",
		WebhookSecret:  "secret1",
	})
	_ = st.AddSource(store.Source{
		Name:           "Source 2",
		Type:           "gitlab",
		DiscordWebhook: "https://discord.com/api/webhooks/2/b",
		WebhookSecret:  "secret2",
	})

	req := httptest.NewRequest("GET", "/api/sources", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []SourceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(resp))
	}

	// Verify secrets are not included
	for i, source := range resp {
		if source.ID == "" {
			t.Errorf("Source %d: expected non-empty ID", i)
		}
		if source.Name == "" {
			t.Errorf("Source %d: expected non-empty Name", i)
		}
	}
}

func TestSourcesHandler_List_ByName(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	// Add sources
	_ = st.AddSource(store.Source{
		Name:           "GitHub Main",
		Type:           "github",
		DiscordWebhook: "https://discord.com/api/webhooks/1/a",
	})
	_ = st.AddSource(store.Source{
		Name:           "GitLab Backup",
		Type:           "gitlab",
		DiscordWebhook: "https://discord.com/api/webhooks/2/b",
	})

	// Query by name
	req := httptest.NewRequest("GET", "/api/sources?name=GitHub+Main", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []SourceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 1 {
		t.Errorf("Expected 1 source, got %d", len(resp))
	}

	if len(resp) > 0 && resp[0].Name != "GitHub Main" {
		t.Errorf("Expected name 'GitHub Main', got %q", resp[0].Name)
	}
}

func TestSourcesHandler_List_ByName_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSourcesHandler(st)

	// Query for non-existent source
	req := httptest.NewRequest("GET", "/api/sources?name=NonExistent", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []SourceResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("Expected empty array, got %d items", len(resp))
	}
}
