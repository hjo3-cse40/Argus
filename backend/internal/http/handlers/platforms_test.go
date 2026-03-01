package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"argus-backend/internal/store"
)

func TestPlatformsHandler_Create_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	reqBody := CreatePlatformRequest{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
		WebhookSecret:  "secret123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response PlatformResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "youtube" {
		t.Errorf("Expected name 'youtube', got '%s'", response.Name)
	}
	if response.DiscordWebhook != "https://discord.com/api/webhooks/123/abc" {
		t.Errorf("Expected webhook URL, got '%s'", response.DiscordWebhook)
	}
	if response.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if response.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

func TestPlatformsHandler_Create_InvalidName(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	reqBody := CreatePlatformRequest{
		Name:           "tiktok",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Validation failed" {
		t.Errorf("Expected 'Validation failed', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_Create_InvalidWebhook(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	reqBody := CreatePlatformRequest{
		Name:           "youtube",
		DiscordWebhook: "https://example.com/webhook",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Validation failed" {
		t.Errorf("Expected 'Validation failed', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_Create_DuplicateName(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Create first platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	// Try to create duplicate
	reqBody := CreatePlatformRequest{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/456/def",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Platform already exists" {
		t.Errorf("Expected 'Platform already exists', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_List(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Add platforms
	platforms := []store.Platform{
		{Name: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123/abc"},
		{Name: "reddit", DiscordWebhook: "https://discord.com/api/webhooks/456/def"},
	}
	for _, p := range platforms {
		if err := st.AddPlatform(p); err != nil {
			t.Fatalf("Failed to add platform: %v", err)
		}
	}

	req := httptest.NewRequest("GET", "/api/platforms", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []PlatformResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 platforms, got %d", len(response))
	}
}

func TestPlatformsHandler_Get_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Add platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	// Get the platform ID
	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found")
	}
	platformID := platforms[0].ID

	req := httptest.NewRequest("GET", "/api/platforms/"+platformID, nil)
	req.SetPathValue("id", platformID)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response PlatformResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != platformID {
		t.Errorf("Expected ID '%s', got '%s'", platformID, response.ID)
	}
	if response.Name != "youtube" {
		t.Errorf("Expected name 'youtube', got '%s'", response.Name)
	}
}

func TestPlatformsHandler_Get_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	req := httptest.NewRequest("GET", "/api/platforms/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Platform not found" {
		t.Errorf("Expected 'Platform not found', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_Update_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Add platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	// Get the platform ID
	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found")
	}
	platformID := platforms[0].ID

	// Update platform
	reqBody := UpdatePlatformRequest{
		DiscordWebhook: "https://discord.com/api/webhooks/999/xyz",
		WebhookSecret:  "newsecret",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/platforms/"+platformID, bytes.NewReader(body))
	req.SetPathValue("id", platformID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response PlatformResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.DiscordWebhook != "https://discord.com/api/webhooks/999/xyz" {
		t.Errorf("Expected updated webhook, got '%s'", response.DiscordWebhook)
	}
}

func TestPlatformsHandler_Update_InvalidWebhook(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Add platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	// Get the platform ID
	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found")
	}
	platformID := platforms[0].ID

	// Try to update with invalid webhook
	reqBody := UpdatePlatformRequest{
		DiscordWebhook: "https://example.com/webhook",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/platforms/"+platformID, bytes.NewReader(body))
	req.SetPathValue("id", platformID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Validation failed" {
		t.Errorf("Expected 'Validation failed', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_Update_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	reqBody := UpdatePlatformRequest{
		DiscordWebhook: "https://discord.com/api/webhooks/999/xyz",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/platforms/nonexistent", bytes.NewReader(body))
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Platform not found" {
		t.Errorf("Expected 'Platform not found', got '%s'", response.Error)
	}
}

func TestPlatformsHandler_Delete_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	// Add platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	// Get the platform ID
	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found")
	}
	platformID := platforms[0].ID

	req := httptest.NewRequest("DELETE", "/api/platforms/"+platformID, nil)
	req.SetPathValue("id", platformID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify platform was deleted
	_, found := st.GetPlatform(platformID)
	if found {
		t.Error("Platform should have been deleted")
	}
}

func TestPlatformsHandler_Delete_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewPlatformsHandler(st)

	req := httptest.NewRequest("DELETE", "/api/platforms/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Platform not found" {
		t.Errorf("Expected 'Platform not found', got '%s'", response.Error)
	}
}
