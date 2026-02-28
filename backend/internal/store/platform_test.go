package store

import (
	"strings"
	"testing"
)

// Test AddPlatform creates platform and returns no error
func TestAddPlatform_Success(t *testing.T) {
	store := NewMemoryStore(100)

	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test",
		WebhookSecret:  "secret123",
	}

	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}

	created := platforms[0]
	if created.Name != "youtube" {
		t.Errorf("Expected name 'youtube', got '%s'", created.Name)
	}
	if created.DiscordWebhook != "https://discord.com/api/webhooks/test" {
		t.Errorf("Expected webhook 'https://discord.com/api/webhooks/test', got '%s'", created.DiscordWebhook)
	}
	if created.WebhookSecret != "secret123" {
		t.Errorf("Expected secret 'secret123', got '%s'", created.WebhookSecret)
	}
	if created.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

// Test AddPlatform with duplicate name returns error (conceptual for MemoryStore)
// Note: MemoryStore doesn't enforce uniqueness, but PostgreSQL will
func TestAddPlatform_DuplicateName(t *testing.T) {
	store := NewMemoryStore(100)

	platform1 := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test1",
	}
	err := store.AddPlatform(platform1)
	if err != nil {
		t.Fatalf("First AddPlatform failed: %v", err)
	}

	platform2 := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test2",
	}
	err = store.AddPlatform(platform2)
	if err == nil {
		t.Fatal("Second AddPlatform should have failed due to duplicate name")
	}

	// Verify error message mentions duplicate name
	if !strings.Contains(err.Error(), "platform name already exists") {
		t.Errorf("Expected duplicate name error, got: %v", err)
	}

	// Verify only one platform exists
	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}
}

// Test ListPlatforms returns all platforms in correct order
func TestListPlatforms_OrderedByName(t *testing.T) {
	store := NewMemoryStore(100)

	// Add platforms in non-alphabetical order
	platforms := []Platform{
		{Name: "x", DiscordWebhook: "https://discord.com/api/webhooks/test1"},
		{Name: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/test2"},
		{Name: "reddit", DiscordWebhook: "https://discord.com/api/webhooks/test3"},
	}

	for _, p := range platforms {
		err := store.AddPlatform(p)
		if err != nil {
			t.Fatalf("AddPlatform failed: %v", err)
		}
	}

	result := store.ListPlatforms()
	if len(result) != 3 {
		t.Fatalf("Expected 3 platforms, got %d", len(result))
	}

	// Verify alphabetical order: reddit, x, youtube
	expectedOrder := []string{"reddit", "x", "youtube"}
	for i, expected := range expectedOrder {
		if result[i].Name != expected {
			t.Errorf("Expected platform[%d] name '%s', got '%s'", i, expected, result[i].Name)
		}
	}
}

// Test GetPlatform retrieves correct platform
func TestGetPlatform_Success(t *testing.T) {
	store := NewMemoryStore(100)

	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test",
	}
	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}
	platformID := platforms[0].ID

	retrieved, found := store.GetPlatform(platformID)
	if !found {
		t.Fatal("Expected to find platform")
	}

	if retrieved.ID != platformID {
		t.Errorf("Expected ID '%s', got '%s'", platformID, retrieved.ID)
	}
	if retrieved.Name != "youtube" {
		t.Errorf("Expected name 'youtube', got '%s'", retrieved.Name)
	}
}

// Test GetPlatform with non-existent ID
func TestGetPlatform_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetPlatform("non-existent-id")
	if found {
		t.Error("Expected not to find platform with non-existent ID")
	}
}

// Test GetPlatformByName retrieves correct platform
func TestGetPlatformByName_Success(t *testing.T) {
	store := NewMemoryStore(100)

	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test",
	}
	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	retrieved, found := store.GetPlatformByName("youtube")
	if !found {
		t.Fatal("Expected to find platform by name")
	}

	if retrieved.Name != "youtube" {
		t.Errorf("Expected name 'youtube', got '%s'", retrieved.Name)
	}
}

// Test GetPlatformByName with non-existent name
func TestGetPlatformByName_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetPlatformByName("non-existent")
	if found {
		t.Error("Expected not to find platform with non-existent name")
	}
}

// Test UpdatePlatform modifies webhook but preserves created_at
func TestUpdatePlatform_PreservesCreatedAt(t *testing.T) {
	store := NewMemoryStore(100)

	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test1",
		WebhookSecret:  "secret1",
	}
	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}
	original := platforms[0]
	originalCreatedAt := original.CreatedAt

	// Update platform
	updated := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test2",
		WebhookSecret:  "secret2",
	}
	err = store.UpdatePlatform(original.ID, updated)
	if err != nil {
		t.Fatalf("UpdatePlatform failed: %v", err)
	}

	// Retrieve updated platform
	result, found := store.GetPlatform(original.ID)
	if !found {
		t.Fatal("Expected to find updated platform")
	}

	// Verify created_at is preserved
	if !result.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("Expected CreatedAt to be preserved: original=%v, updated=%v", originalCreatedAt, result.CreatedAt)
	}

	// Verify other fields are updated
	if result.DiscordWebhook != "https://discord.com/api/webhooks/test2" {
		t.Errorf("Expected webhook 'https://discord.com/api/webhooks/test2', got '%s'", result.DiscordWebhook)
	}
	if result.WebhookSecret != "secret2" {
		t.Errorf("Expected secret 'secret2', got '%s'", result.WebhookSecret)
	}
}

// Test DeletePlatform removes platform and cascades to subsources
func TestDeletePlatform_CascadesToSubsources(t *testing.T) {
	store := NewMemoryStore(100)

	// Create platform
	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test",
	}
	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}
	platformID := platforms[0].ID

	// Create subsources
	subsource1 := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource1)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsource2 := Subsource{
		PlatformID: platformID,
		Name:       "NFL",
		Identifier: "UCyyy",
	}
	err = store.AddSubsource(subsource2)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	// Verify subsources exist
	subsources := store.ListAllSubsources()
	if len(subsources) != 2 {
		t.Fatalf("Expected 2 subsources, got %d", len(subsources))
	}

	// Delete platform
	err = store.DeletePlatform(platformID)
	if err != nil {
		t.Fatalf("DeletePlatform failed: %v", err)
	}

	// Verify platform is deleted
	platforms = store.ListPlatforms()
	if len(platforms) != 0 {
		t.Errorf("Expected 0 platforms after deletion, got %d", len(platforms))
	}

	// Verify subsources are deleted (cascade)
	subsources = store.ListAllSubsources()
	if len(subsources) != 0 {
		t.Errorf("Expected 0 subsources after platform deletion (cascade), got %d", len(subsources))
	}
}

// Test DeletePlatform with non-existent ID
func TestDeletePlatform_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	err := store.DeletePlatform("non-existent-id")
	// MemoryStore doesn't return error for non-existent ID
	if err != nil {
		t.Errorf("Expected no error for non-existent ID, got: %v", err)
	}
}
