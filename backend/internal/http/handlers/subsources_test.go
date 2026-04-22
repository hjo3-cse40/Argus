package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"argus-backend/internal/store"
)

func TestSubsourcesHandler_Create_Success(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform first
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found after adding")
	}
	platformID := platforms[0].ID

	reqBody := CreateSubsourceRequest{
		Name:       "NBA",
		Identifier: "UCxxx",
		URL:        "https://youtube.com/channel/UCxxx",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms/"+platformID+"/subsources", bytes.NewReader(body))
	req.SetPathValue("platform_id", platformID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response SubsourceResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "NBA" {
		t.Errorf("Expected name 'NBA', got '%s'", response.Name)
	}
	if response.Identifier != "UCxxx" {
		t.Errorf("Expected identifier 'UCxxx', got '%s'", response.Identifier)
	}
	if response.PlatformID != platformID {
		t.Errorf("Expected platform_id '%s', got '%s'", platformID, response.PlatformID)
	}
	if response.PlatformName != "youtube" {
		t.Errorf("Expected platform_name 'youtube', got '%s'", response.PlatformName)
	}
	if response.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if response.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

func TestSubsourcesHandler_Create_InvalidPlatformID(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	reqBody := CreateSubsourceRequest{
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms/nonexistent/subsources", bytes.NewReader(body))
	req.SetPathValue("platform_id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	bodyBytes := w.Body.Bytes()
	var response ErrorResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// The store validates platform existence and returns "platform not found" error
	// The handler catches this and returns it with appropriate details
	if response.Error != "Platform not found" {
		t.Errorf("Expected 'Platform not found', got '%s'", response.Error)
	}
	if len(response.Details) == 0 || response.Details[0] != "platform_id does not reference an existing platform" {
		t.Errorf("Expected details about invalid platform_id, got %v", response.Details)
	}
}

func TestSubsourcesHandler_Create_EmptyName(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform first
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	reqBody := CreateSubsourceRequest{
		Name:       "",
		Identifier: "UCxxx",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms/"+platformID+"/subsources", bytes.NewReader(body))
	req.SetPathValue("platform_id", platformID)
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

func TestSubsourcesHandler_Create_EmptyIdentifier(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform first
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	reqBody := CreateSubsourceRequest{
		Name:       "NBA",
		Identifier: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms/"+platformID+"/subsources", bytes.NewReader(body))
	req.SetPathValue("platform_id", platformID)
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

func TestSubsourcesHandler_Create_DuplicateIdentifier(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform first
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	// Create first subsource
	subsource := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	if err := st.AddSubsource(subsource); err != nil {
		t.Fatalf("Failed to add subsource: %v", err)
	}

	// Try to create duplicate
	reqBody := CreateSubsourceRequest{
		Name:       "NBA2",
		Identifier: "UCxxx",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/platforms/"+platformID+"/subsources", bytes.NewReader(body))
	req.SetPathValue("platform_id", platformID)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != "Subsource already exists" {
		t.Errorf("Expected 'Subsource already exists', got '%s'", response.Error)
	}
}

func TestSubsourcesHandler_ListByPlatform(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	// Create subsources
	subsource1 := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	subsource2 := store.Subsource{
		PlatformID: platformID,
		Name:       "NFL",
		Identifier: "UCyyy",
	}
	if err := st.AddSubsource(subsource1); err != nil {
		t.Fatalf("Failed to add subsource1: %v", err)
	}
	if err := st.AddSubsource(subsource2); err != nil {
		t.Fatalf("Failed to add subsource2: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/platforms/"+platformID+"/subsources", nil)
	req.SetPathValue("platform_id", platformID)
	w := httptest.NewRecorder()

	handler.ListByPlatform(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var responses []SubsourceResponse
	if err := json.NewDecoder(w.Body).Decode(&responses); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 subsources, got %d", len(responses))
	}

	// Verify platform_name is included
	for _, resp := range responses {
		if resp.PlatformName != "youtube" {
			t.Errorf("Expected platform_name 'youtube', got '%s'", resp.PlatformName)
		}
	}
}

func TestSubsourcesHandler_Get(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	// Create subsource
	subsource := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	if err := st.AddSubsource(subsource); err != nil {
		t.Fatalf("Failed to add subsource: %v", err)
	}

	subsources := st.ListSubsources(platformID)
	subsourceID := subsources[0].ID

	req := httptest.NewRequest("GET", "/api/subsources/"+subsourceID, nil)
	req.SetPathValue("id", subsourceID)
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response SubsourceResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "NBA" {
		t.Errorf("Expected name 'NBA', got '%s'", response.Name)
	}
	if response.PlatformName != "youtube" {
		t.Errorf("Expected platform_name 'youtube', got '%s'", response.PlatformName)
	}
}

func TestSubsourcesHandler_Get_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	req := httptest.NewRequest("GET", "/api/subsources/nonexistent", nil)
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

	if response.Error != "Subsource not found" {
		t.Errorf("Expected 'Subsource not found', got '%s'", response.Error)
	}
}

func TestSubsourcesHandler_Update(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	// Create subsource
	subsource := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	if err := st.AddSubsource(subsource); err != nil {
		t.Fatalf("Failed to add subsource: %v", err)
	}

	subsources := st.ListSubsources(platformID)
	subsourceID := subsources[0].ID

	reqBody := UpdateSubsourceRequest{
		Name:       "NBA Updated",
		Identifier: "UCxxx",
		URL:        "https://youtube.com/channel/UCxxx",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/subsources/"+subsourceID, bytes.NewReader(body))
	req.SetPathValue("id", subsourceID)
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response SubsourceResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Name != "NBA Updated" {
		t.Errorf("Expected name 'NBA Updated', got '%s'", response.Name)
	}
	if response.Identifier != "UCxxx" {
		t.Errorf("Expected identifier to remain 'UCxxx', got '%s'", response.Identifier)
	}
	if response.PlatformID != platformID {
		t.Errorf("Expected platform_id to remain '%s', got '%s'", platformID, response.PlatformID)
	}
}

func TestSubsourcesHandler_Update_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	reqBody := UpdateSubsourceRequest{
		Name: "NBA Updated",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/subsources/nonexistent", bytes.NewReader(body))
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

	if response.Error != "Subsource not found" {
		t.Errorf("Expected 'Subsource not found', got '%s'", response.Error)
	}
}

func TestSubsourcesHandler_Delete(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	// Create a platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	platformID := platforms[0].ID

	// Create subsource
	subsource := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	if err := st.AddSubsource(subsource); err != nil {
		t.Fatalf("Failed to add subsource: %v", err)
	}

	subsources := st.ListSubsources(platformID)
	subsourceID := subsources[0].ID

	req := httptest.NewRequest("DELETE", "/api/subsources/"+subsourceID, nil)
	req.SetPathValue("id", subsourceID)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify subsource was deleted
	_, found := st.GetSubsource(subsourceID)
	if found {
		t.Error("Expected subsource to be deleted")
	}
}

func TestSubsourcesHandler_Delete_NotFound(t *testing.T) {
	st := store.NewMemoryStore(100)
	handler := NewSubsourcesHandler(st)

	req := httptest.NewRequest("DELETE", "/api/subsources/nonexistent", nil)
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

	if response.Error != "Subsource not found" {
		t.Errorf("Expected 'Subsource not found', got '%s'", response.Error)
	}
}
