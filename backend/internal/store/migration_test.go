package store

import (
	"testing"
	"time"
)

// Test migration creates platforms from unique source types
func TestMigrationCreatesPlatformsFromUniqueSourceTypes(t *testing.T) {
	store := NewMemoryStore(100)

	// Create sources with different types
	sources := []Source{
		{Name: "NBA YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123", WebhookSecret: "secret1"},
		{Name: "NFL YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123", WebhookSecret: "secret1"},
		{Name: "r/nba", Type: "reddit", DiscordWebhook: "https://discord.com/api/webhooks/456", WebhookSecret: "secret2"},
		{Name: "ElonMusk", Type: "x", DiscordWebhook: "https://discord.com/api/webhooks/789", WebhookSecret: "secret3"},
	}

	for _, source := range sources {
		err := store.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add source: %v", err)
		}
	}

	// Run migration
	err := MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify platforms created
	platforms := store.ListPlatforms()
	if len(platforms) != 3 {
		t.Errorf("Expected 3 platforms, got %d", len(platforms))
	}

	// Verify platform names
	platformNames := make(map[string]bool)
	for _, platform := range platforms {
		platformNames[platform.Name] = true
	}

	expectedNames := []string{"youtube", "reddit", "x"}
	for _, name := range expectedNames {
		if !platformNames[name] {
			t.Errorf("Expected platform %s not found", name)
		}
	}
}

// Test migration creates subsources from sources
func TestMigrationCreatesSubsourcesFromSources(t *testing.T) {
	store := NewMemoryStore(100)

	// Create sources
	sources := []Source{
		{Name: "NBA YouTube", Type: "youtube", RepositoryURL: "UCxxx", DiscordWebhook: "https://discord.com/api/webhooks/123"},
		{Name: "r/nba", Type: "reddit", RepositoryURL: "nba", DiscordWebhook: "https://discord.com/api/webhooks/456"},
	}

	for _, source := range sources {
		err := store.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add source: %v", err)
		}
	}

	// Run migration
	err := MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify subsources created
	subsources := store.ListAllSubsources()
	if len(subsources) != 2 {
		t.Errorf("Expected 2 subsources, got %d", len(subsources))
	}

	// Verify subsource data
	subsourceNames := make(map[string]SubsourceWithPlatform)
	for _, subsource := range subsources {
		subsourceNames[subsource.Name] = subsource
	}

	// Check NBA YouTube subsource
	nbaYT, exists := subsourceNames["NBA YouTube"]
	if !exists {
		t.Error("NBA YouTube subsource not found")
	} else {
		if nbaYT.PlatformName != "youtube" {
			t.Errorf("Expected platform name 'youtube', got '%s'", nbaYT.PlatformName)
		}
		// URL should be auto-generated, not the raw identifier
		expectedURL := "https://youtube.com/channel/UCxxx"
		if nbaYT.URL != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, nbaYT.URL)
		}
	}

	// Check r/nba subsource
	rnba, exists := subsourceNames["r/nba"]
	if !exists {
		t.Error("r/nba subsource not found")
	} else {
		if rnba.PlatformName != "reddit" {
			t.Errorf("Expected platform name 'reddit', got '%s'", rnba.PlatformName)
		}
		// URL should be auto-generated, not the raw identifier
		expectedURL := "https://reddit.com/r/nba"
		if rnba.URL != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, rnba.URL)
		}
	}
}

// Test migration preserves timestamps
func TestMigrationPreservesTimestamps(t *testing.T) {
	store := NewMemoryStore(100)

	// Create source with specific timestamp
	now := time.Now().UTC()
	source := Source{
		Name:           "Test Source",
		Type:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123",
		CreatedAt:      now,
	}

	err := store.AddSource(source)
	if err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Run migration
	err = MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify platform timestamp
	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Fatalf("Expected 1 platform, got %d", len(platforms))
	}
	if !platforms[0].CreatedAt.Equal(now) {
		t.Errorf("Platform timestamp not preserved. Expected %v, got %v", now, platforms[0].CreatedAt)
	}

	// Verify subsource timestamp
	subsources := store.ListAllSubsources()
	if len(subsources) != 1 {
		t.Fatalf("Expected 1 subsource, got %d", len(subsources))
	}
	if !subsources[0].CreatedAt.Equal(now) {
		t.Errorf("Subsource timestamp not preserved. Expected %v, got %v", now, subsources[0].CreatedAt)
	}
}

// Test migration fails on inconsistent webhooks with descriptive error
func TestMigrationFailsOnInconsistentWebhooks(t *testing.T) {
	store := NewMemoryStore(100)

	// Create sources with same type but different webhooks
	sources := []Source{
		{Name: "NBA YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123"},
		{Name: "NFL YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/456"},
	}

	for _, source := range sources {
		err := store.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add source: %v", err)
		}
	}

	// Run migration - should fail
	err := MigrateFlatToHierarchical(store)
	if err == nil {
		t.Fatal("Expected migration to fail due to inconsistent webhooks")
	}

	// Verify error message is descriptive
	expectedError := "migration failed: inconsistent webhooks detected for platform 'youtube'"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Verify no platforms were created
	platforms := store.ListPlatforms()
	if len(platforms) != 0 {
		t.Errorf("Expected 0 platforms after failed migration, got %d", len(platforms))
	}

	// Verify no subsources were created
	subsources := store.ListAllSubsources()
	if len(subsources) != 0 {
		t.Errorf("Expected 0 subsources after failed migration, got %d", len(subsources))
	}
}

// Test migration is idempotent (can run multiple times safely)
func TestMigrationIsIdempotent(t *testing.T) {
	store := NewMemoryStore(100)

	// Create source
	source := Source{
		Name:           "Test Source",
		Type:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123",
	}

	err := store.AddSource(source)
	if err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Run migration first time
	err = MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("First migration failed: %v", err)
	}

	// Verify initial state
	platforms := store.ListPlatforms()
	subsources := store.ListAllSubsources()
	if len(platforms) != 1 || len(subsources) != 1 {
		t.Fatalf("Expected 1 platform and 1 subsource after first migration")
	}

	// Run migration second time - should not fail or create duplicates
	err = MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Second migration failed: %v", err)
	}

	// Verify state unchanged
	platforms2 := store.ListPlatforms()
	subsources2 := store.ListAllSubsources()
	if len(platforms2) != 1 || len(subsources2) != 1 {
		t.Errorf("Expected 1 platform and 1 subsource after second migration, got %d platforms and %d subsources",
			len(platforms2), len(subsources2))
	}
}

// Test migration handles empty sources table
func TestMigrationHandlesEmptySourcesTable(t *testing.T) {
	store := NewMemoryStore(100)

	// Run migration on empty store - should not fail
	err := MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed on empty store: %v", err)
	}

	// Verify no platforms or subsources created
	platforms := store.ListPlatforms()
	subsources := store.ListAllSubsources()
	if len(platforms) != 0 || len(subsources) != 0 {
		t.Errorf("Expected 0 platforms and 0 subsources for empty migration, got %d platforms and %d subsources",
			len(platforms), len(subsources))
	}
}

// Test migration creates multiple subsources for shared platform
func TestMigrationCreatesMultipleSubsourcesForSharedPlatform(t *testing.T) {
	store := NewMemoryStore(100)

	// Create multiple sources with same type and webhook
	sources := []Source{
		{Name: "NBA YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123", WebhookSecret: "secret"},
		{Name: "NFL YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123", WebhookSecret: "secret"},
		{Name: "MLB YouTube", Type: "youtube", DiscordWebhook: "https://discord.com/api/webhooks/123", WebhookSecret: "secret"},
	}

	for _, source := range sources {
		err := store.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add source: %v", err)
		}
	}

	// Run migration
	err := MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify exactly one platform created
	platforms := store.ListPlatforms()
	if len(platforms) != 1 {
		t.Errorf("Expected 1 platform, got %d", len(platforms))
	}

	platform := platforms[0]
	if platform.Name != "youtube" {
		t.Errorf("Expected platform name 'youtube', got '%s'", platform.Name)
	}
	if platform.DiscordWebhook != "https://discord.com/api/webhooks/123" {
		t.Errorf("Expected webhook 'https://discord.com/api/webhooks/123', got '%s'", platform.DiscordWebhook)
	}

	// Verify multiple subsources created, all referencing the same platform
	subsources := store.ListAllSubsources()
	if len(subsources) != 3 {
		t.Errorf("Expected 3 subsources, got %d", len(subsources))
	}

	for _, subsource := range subsources {
		if subsource.PlatformID != platform.ID {
			t.Errorf("Subsource %s has wrong platform_id. Expected %s, got %s",
				subsource.Name, platform.ID, subsource.PlatformID)
		}
		if subsource.PlatformName != "youtube" {
			t.Errorf("Subsource %s has wrong platform name. Expected 'youtube', got '%s'",
				subsource.Name, subsource.PlatformName)
		}
	}

	// Verify subsource names match original source names
	expectedNames := []string{"NBA YouTube", "NFL YouTube", "MLB YouTube"}
	actualNames := make(map[string]bool)
	for _, subsource := range subsources {
		actualNames[subsource.Name] = true
	}

	for _, expectedName := range expectedNames {
		if !actualNames[expectedName] {
			t.Errorf("Expected subsource name '%s' not found", expectedName)
		}
	}
}

// Test migration uses repository_url as identifier when available
func TestMigrationUsesRepositoryURLAsIdentifier(t *testing.T) {
	store := NewMemoryStore(100)

	// Create sources with and without repository_url
	sources := []Source{
		{Name: "NBA YouTube", Type: "youtube", RepositoryURL: "UCxxx123", DiscordWebhook: "https://discord.com/api/webhooks/123"},
		{Name: "NFL YouTube", Type: "youtube", RepositoryURL: "", DiscordWebhook: "https://discord.com/api/webhooks/123"},
	}

	for _, source := range sources {
		err := store.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add source: %v", err)
		}
	}

	// Run migration
	err := MigrateFlatToHierarchical(store)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify subsources
	subsources := store.ListAllSubsources()
	if len(subsources) != 2 {
		t.Fatalf("Expected 2 subsources, got %d", len(subsources))
	}

	// Find subsources by name
	var nbaSubsource, nflSubsource SubsourceWithPlatform
	for _, subsource := range subsources {
		switch subsource.Name {
		case "NBA YouTube":
			nbaSubsource = subsource
		case "NFL YouTube":
			nflSubsource = subsource
		}
	}

	// Verify NBA YouTube uses repository_url as identifier
	if nbaSubsource.Identifier != "UCxxx123" {
		t.Errorf("Expected NBA YouTube identifier 'UCxxx123', got '%s'", nbaSubsource.Identifier)
	}
	// URL should be auto-generated from platform and identifier
	expectedNBAURL := "https://youtube.com/channel/UCxxx123"
	if nbaSubsource.URL != expectedNBAURL {
		t.Errorf("Expected NBA YouTube URL '%s', got '%s'", expectedNBAURL, nbaSubsource.URL)
	}

	// Verify NFL YouTube uses name as identifier (fallback)
	if nflSubsource.Identifier != "NFL YouTube" {
		t.Errorf("Expected NFL YouTube identifier 'NFL YouTube', got '%s'", nflSubsource.Identifier)
	}
	// URL should be auto-generated from platform and identifier
	expectedNFLURL := "https://youtube.com/channel/NFL YouTube"
	if nflSubsource.URL != expectedNFLURL {
		t.Errorf("Expected NFL YouTube URL '%s', got '%s'", expectedNFLURL, nflSubsource.URL)
	}
}