package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"argus-backend/internal/store"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: source-registration-ui, Property 3: Secret Exclusion from List Responses
// For any source configuration created with a webhook secret,
// when retrieving the source list via GET /api/sources,
// the response should not contain the webhook_secret field for any source.
func TestProperty_SecretExclusionFromListResponses(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("list response never contains webhook_secret", prop.ForAll(
		func(name, sourceType, webhook, secret string) bool {
			// Create store and handler
			st := store.NewMemoryStore(100)
			handler := NewSourcesHandler(st)

			// Add source with secret
			source := store.Source{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: webhook,
				WebhookSecret:  secret,
			}
			_ = st.AddSource(source)

			// Make GET request
			req := httptest.NewRequest("GET", "/api/sources", nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			// Check response
			if w.Code != http.StatusOK {
				return false
			}

			// Parse response
			var responses []map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&responses); err != nil {
				return false
			}

			if len(responses) != 1 {
				return false
			}

			// Verify webhook_secret is not in response
			_, hasSecret := responses[0]["webhook_secret"]
			if hasSecret {
				return false // Secret should not be present
			}

			// Verify other fields are present
			if _, ok := responses[0]["id"]; !ok {
				return false
			}
			if _, ok := responses[0]["name"]; !ok {
				return false
			}
			if _, ok := responses[0]["type"]; !ok {
				return false
			}
			if _, ok := responses[0]["discord_webhook"]; !ok {
				return false
			}
			if _, ok := responses[0]["created_at"]; !ok {
				return false
			}

			return true
		},
		genValidSourceName(),
		genValidSourceType(),
		genValidDiscordWebhook(),
		gen.AlphaString(), // Any secret
	))

	properties.TestingRun(t)
}

// Feature: source-registration-ui, Property 6: Response Structure Completeness
// For any source returned by GET /api/sources, the response should include
// id, name, type, discord_webhook, and created_at fields,
// and should not include webhook_secret.
func TestProperty_ResponseStructureCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("response contains all required fields", prop.ForAll(
		func(name, sourceType, webhook string) bool {
			// Create store and handler
			st := store.NewMemoryStore(100)
			handler := NewSourcesHandler(st)

			// Add source
			source := store.Source{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: webhook,
			}
			_ = st.AddSource(source)

			// Make GET request
			req := httptest.NewRequest("GET", "/api/sources", nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			// Parse response
			var responses []SourceResponse
			if err := json.NewDecoder(w.Body).Decode(&responses); err != nil {
				return false
			}

			if len(responses) != 1 {
				return false
			}

			resp := responses[0]

			// Verify all required fields are present and non-empty
			if resp.ID == "" {
				return false
			}
			if resp.Name != name {
				return false
			}
			if resp.Type != sourceType {
				return false
			}
			if resp.DiscordWebhook != webhook {
				return false
			}
			if resp.CreatedAt.IsZero() {
				return false
			}

			return true
		},
		genValidSourceName(),
		genValidSourceType(),
		genValidDiscordWebhook(),
	))

	properties.TestingRun(t)
}
