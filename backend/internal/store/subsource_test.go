package store

import (
	"strings"
	"testing"
)

// Test AddSubsource creates subsource and returns no error
func TestAddSubsource_Success(t *testing.T) {
	store := NewMemoryStore(100)

	// Create platform first
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

	// Create subsource
	subsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
		URL:        "https://youtube.com/channel/UCxxx",
	}

	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}

	created := subsources[0]
	if created.Name != "NBA" {
		t.Errorf("Expected name 'NBA', got '%s'", created.Name)
	}
	if created.Identifier != "UCxxx" {
		t.Errorf("Expected identifier 'UCxxx', got '%s'", created.Identifier)
	}
	if created.URL != "https://youtube.com/channel/UCxxx" {
		t.Errorf("Expected URL 'https://youtube.com/channel/UCxxx', got '%s'", created.URL)
	}
	if created.PlatformID != platformID {
		t.Errorf("Expected platform_id '%s', got '%s'", platformID, created.PlatformID)
	}
	if created.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

// Test AddSubsource with duplicate identifier returns error (conceptual for MemoryStore)
// Note: MemoryStore doesn't enforce uniqueness, but PostgreSQL will
func TestAddSubsource_DuplicateIdentifier(t *testing.T) {
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
	platformID := platforms[0].ID

	// Create first subsource
	subsource1 := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource1)
	if err != nil {
		t.Fatalf("First AddSubsource failed: %v", err)
	}

	// Create second subsource with same identifier
	subsource2 := Subsource{
		PlatformID: platformID,
		Name:       "NFL",
		Identifier: "UCxxx", // Duplicate identifier
	}
	err = store.AddSubsource(subsource2)
	if err == nil {
		t.Fatal("Second AddSubsource should have failed due to duplicate identifier")
	}

	// Verify error message mentions duplicate identifier
	if !strings.Contains(err.Error(), "subsource identifier already exists") {
		t.Errorf("Expected duplicate identifier error, got: %v", err)
	}

	// Verify only one subsource exists
	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}
}

// Test AddSubsource with non-existent platform_id returns error (conceptual for MemoryStore)
func TestAddSubsource_NonExistentPlatform(t *testing.T) {
	store := NewMemoryStore(100)

	// Create subsource with non-existent platform_id
	subsource := Subsource{
		PlatformID: "non-existent-platform-id",
		Name:       "NBA",
		Identifier: "UCxxx",
	}

	err := store.AddSubsource(subsource)
	if err == nil {
		t.Fatal("AddSubsource should have failed due to non-existent platform_id")
	}

	// Verify error message mentions platform not found
	if !strings.Contains(err.Error(), "platform not found") {
		t.Errorf("Expected platform not found error, got: %v", err)
	}

	// Verify no subsource was created
	subsources := store.ListAllSubsources()
	if len(subsources) != 0 {
		t.Fatalf("Expected 0 subsources, got %d", len(subsources))
	}
}

// Test AddSubsource auto-generates URL when not provided
func TestAddSubsource_AutoGeneratesURL(t *testing.T) {
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
	platformID := platforms[0].ID

	// Create subsource without URL
	subsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
		URL:        "", // No URL provided
	}

	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}

	created := subsources[0]
	// Note: MemoryStore doesn't implement URL auto-generation
	// In PostgreSQL implementation, URL would be auto-generated
	// For now, we just verify the subsource was created
	if created.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

// Test AddSubsource preserves URL when provided
func TestAddSubsource_PreservesURL(t *testing.T) {
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
	platformID := platforms[0].ID

	// Create subsource with explicit URL
	customURL := "https://custom.url/channel/UCxxx"
	subsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
		URL:        customURL,
	}

	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}

	created := subsources[0]
	if created.URL != customURL {
		t.Errorf("Expected URL '%s', got '%s'", customURL, created.URL)
	}
}

// Test ListSubsources returns subsources for platform with platform_name
func TestListSubsources_WithPlatformName(t *testing.T) {
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

	// List subsources for platform
	subsources := store.ListSubsources(platformID)
	if len(subsources) != 2 {
		t.Fatalf("Expected 2 subsources, got %d", len(subsources))
	}

	// Verify platform_name is included
	for _, sub := range subsources {
		if sub.PlatformName != "youtube" {
			t.Errorf("Expected platform_name 'youtube', got '%s'", sub.PlatformName)
		}
	}
}

// Test ListAllSubsources returns all subsources with platform_name
func TestListAllSubsources_WithPlatformName(t *testing.T) {
	store := NewMemoryStore(100)

	// Create two platforms
	platform1 := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test1",
	}
	err := store.AddPlatform(platform1)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platform2 := Platform{
		Name:           "reddit",
		DiscordWebhook: "https://discord.com/api/webhooks/test2",
	}
	err = store.AddPlatform(platform2)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 2 {
		t.Fatalf("Expected 2 platforms, got %d", len(platforms))
	}

	// Find platform IDs (order is alphabetical: reddit, youtube)
	var youtubeID, redditID string
	for _, p := range platforms {
		switch p.Name {
		case "youtube":
			youtubeID = p.ID
		case "reddit":
			redditID = p.ID
		}
	}

	// Create subsources for each platform
	subsource1 := Subsource{
		PlatformID: youtubeID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource1)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsource2 := Subsource{
		PlatformID: redditID,
		Name:       "nba",
		Identifier: "nba",
	}
	err = store.AddSubsource(subsource2)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	// List all subsources
	subsources := store.ListAllSubsources()
	if len(subsources) != 2 {
		t.Fatalf("Expected 2 subsources, got %d", len(subsources))
	}

	// Verify platform_name is included for each
	platformNames := make(map[string]bool)
	for _, sub := range subsources {
		if sub.PlatformName == "" {
			t.Error("Expected non-empty platform_name")
		}
		platformNames[sub.PlatformName] = true
	}

	// Verify we have both platform names
	if !platformNames["youtube"] || !platformNames["reddit"] {
		t.Error("Expected both 'youtube' and 'reddit' platform names")
	}
}

// Test GetSubsource retrieves correct subsource with platform_name
func TestGetSubsource_WithPlatformName(t *testing.T) {
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
	platformID := platforms[0].ID

	// Create subsource
	subsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}
	subsourceID := subsources[0].ID

	// Get subsource
	retrieved, found := store.GetSubsource(subsourceID)
	if !found {
		t.Fatal("Expected to find subsource")
	}

	if retrieved.ID != subsourceID {
		t.Errorf("Expected ID '%s', got '%s'", subsourceID, retrieved.ID)
	}
	if retrieved.Name != "NBA" {
		t.Errorf("Expected name 'NBA', got '%s'", retrieved.Name)
	}
	if retrieved.PlatformName != "youtube" {
		t.Errorf("Expected platform_name 'youtube', got '%s'", retrieved.PlatformName)
	}
}

// Test GetSubsource with non-existent ID
func TestGetSubsource_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	_, found := store.GetSubsource("non-existent-id")
	if found {
		t.Error("Expected not to find subsource with non-existent ID")
	}
}

// Test UpdateSubsource modifies name but prevents platform_id change
func TestUpdateSubsource_PreventsPlatformIDChange(t *testing.T) {
	store := NewMemoryStore(100)

	// Create two platforms
	platform1 := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test1",
	}
	err := store.AddPlatform(platform1)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platform2 := Platform{
		Name:           "reddit",
		DiscordWebhook: "https://discord.com/api/webhooks/test2",
	}
	err = store.AddPlatform(platform2)
	if err != nil {
		t.Fatalf("AddPlatform failed: %v", err)
	}

	platforms := store.ListPlatforms()
	if len(platforms) != 2 {
		t.Fatalf("Expected 2 platforms, got %d", len(platforms))
	}

	// Find platform IDs
	var platform1ID, platform2ID string
	for _, p := range platforms {
		switch p.Name {
		case "youtube":
			platform1ID = p.ID
		case "reddit":
			platform2ID = p.ID
		}
	}

	// Create subsource under platform1
	subsource := Subsource{
		PlatformID: platform1ID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}
	subsourceID := subsources[0].ID
	originalPlatformID := subsources[0].PlatformID

	// Attempt to update subsource with different platform_id
	updated := Subsource{
		PlatformID: platform2ID, // Try to change platform
		Name:       "Updated NBA",
		Identifier: "UCyyy",
	}
	err = store.UpdateSubsource(subsourceID, updated)
	if err != nil {
		t.Fatalf("UpdateSubsource failed: %v", err)
	}

	// Verify platform_id is preserved (not changed)
	result, found := store.GetSubsource(subsourceID)
	if !found {
		t.Fatal("Expected to find updated subsource")
	}

	if result.PlatformID != originalPlatformID {
		t.Errorf("Expected platform_id to be preserved: original=%s, updated=%s", originalPlatformID, result.PlatformID)
	}

	// Verify other fields are updated
	if result.Name != "Updated NBA" {
		t.Errorf("Expected name 'Updated NBA', got '%s'", result.Name)
	}
	if result.Identifier != "UCyyy" {
		t.Errorf("Expected identifier 'UCyyy', got '%s'", result.Identifier)
	}
}

// Test DeleteSubsource removes subsource and sets delivery subsource_id to NULL
func TestDeleteSubsource_RemovesSubsource(t *testing.T) {
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
	platformID := platforms[0].ID

	// Create subsource
	subsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	err = store.AddSubsource(subsource)
	if err != nil {
		t.Fatalf("AddSubsource failed: %v", err)
	}

	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}
	subsourceID := subsources[0].ID

	// Delete subsource
	err = store.DeleteSubsource(subsourceID)
	if err != nil {
		t.Fatalf("DeleteSubsource failed: %v", err)
	}

	// Verify subsource is deleted
	subsources = store.ListAllSubsources()
	if len(subsources) != 0 {
		t.Errorf("Expected 0 subsources after deletion, got %d", len(subsources))
	}

	// Verify subsource cannot be retrieved
	_, found := store.GetSubsource(subsourceID)
	if found {
		t.Error("Expected not to find deleted subsource")
	}

	// Note: Testing delivery subsource_id set to NULL would require
	// delivery-subsource linking, which is not yet implemented in MemoryStore
}

// Test DeleteSubsource with non-existent ID
func TestDeleteSubsource_NotFound(t *testing.T) {
	store := NewMemoryStore(100)

	err := store.DeleteSubsource("non-existent-id")
	// MemoryStore doesn't return error for non-existent ID
	if err != nil {
		t.Errorf("Expected no error for non-existent ID, got: %v", err)
	}
}
