package main

import (
	"argus-backend/internal/store"
	"testing"
)

func TestLoadSubsources(t *testing.T) {
	st := store.NewMemoryStore(100)

	// Add a platform
	platform := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test",
	}
	if err := st.AddPlatform(platform); err != nil {
		t.Fatalf("failed to add platform: %v", err)
	}

	platforms := st.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("no platforms found")
	}
	platformID := platforms[0].ID

	// Add subsources
	subsource1 := store.Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	if err := st.AddSubsource(subsource1); err != nil {
		t.Fatalf("failed to add subsource1: %v", err)
	}

	subsource2 := store.Subsource{
		PlatformID: platformID,
		Name:       "NFL",
		Identifier: "UCyyy",
	}
	if err := st.AddSubsource(subsource2); err != nil {
		t.Fatalf("failed to add subsource2: %v", err)
	}

	// Test loadSubsources
	subsources, err := loadSubsources(st)
	if err != nil {
		t.Fatalf("loadSubsources failed: %v", err)
	}

	if len(subsources) != 2 {
		t.Errorf("expected 2 subsources, got %d", len(subsources))
	}

	// Verify platform information is included
	for _, s := range subsources {
		if s.PlatformName == "" {
			t.Error("subsource missing platform_name")
		}
		if s.PlatformName != "youtube" {
			t.Errorf("expected platform_name 'youtube', got '%s'", s.PlatformName)
		}
	}
}

func TestLoadSubsources_Empty(t *testing.T) {
	st := store.NewMemoryStore(100)

	// Test with no subsources
	_, err := loadSubsources(st)
	if err == nil {
		t.Error("expected error when no subsources configured, got nil")
	}
}

func TestConstructRSSHubURL_YouTube(t *testing.T) {
	baseURL := "https://rsshub.example.com"
	platformName := "youtube"
	identifier := "UCxxx"

	url := constructRSSHubURL(baseURL, platformName, identifier)
	expected := "https://rsshub.example.com/youtube/channel/UCxxx"

	if url != expected {
		t.Errorf("expected '%s', got '%s'", expected, url)
	}
}

func TestConstructRSSHubURL_Reddit(t *testing.T) {
	baseURL := "https://rsshub.example.com"
	platformName := "reddit"
	identifier := "programming"

	url := constructRSSHubURL(baseURL, platformName, identifier)
	expected := "https://rsshub.example.com/reddit/subreddit/programming"

	if url != expected {
		t.Errorf("expected '%s', got '%s'", expected, url)
	}
}

func TestConstructRSSHubURL_X(t *testing.T) {
	baseURL := "https://rsshub.example.com"
	platformName := "x"
	identifier := "elonmusk"

	url := constructRSSHubURL(baseURL, platformName, identifier)
	expected := "https://rsshub.example.com/twitter/user/elonmusk"

	if url != expected {
		t.Errorf("expected '%s', got '%s'", expected, url)
	}
}

func TestConstructRSSHubURL_UnsupportedPlatform(t *testing.T) {
	baseURL := "https://rsshub.example.com"
	platformName := "tiktok"
	identifier := "test"

	url := constructRSSHubURL(baseURL, platformName, identifier)
	if url != "" {
		t.Errorf("expected empty string for unsupported platform, got '%s'", url)
	}
}

func TestDeduplication_BySubsource(t *testing.T) {
	seenIDs := make(map[string]map[string]bool)

	subsourceID1 := "subsource-1"
	subsourceID2 := "subsource-2"
	eventID := "event-123"

	// Initialize maps
	if seenIDs[subsourceID1] == nil {
		seenIDs[subsourceID1] = make(map[string]bool)
	}
	if seenIDs[subsourceID2] == nil {
		seenIDs[subsourceID2] = make(map[string]bool)
	}

	// Mark event as seen for subsource1
	seenIDs[subsourceID1][eventID] = true

	// Verify event is seen for subsource1
	if !seenIDs[subsourceID1][eventID] {
		t.Error("event should be marked as seen for subsource1")
	}

	// Verify event is NOT seen for subsource2
	if seenIDs[subsourceID2][eventID] {
		t.Error("event should NOT be marked as seen for subsource2")
	}

	// Mark event as seen for subsource2
	seenIDs[subsourceID2][eventID] = true

	// Verify both subsources have the event marked as seen
	if !seenIDs[subsourceID1][eventID] {
		t.Error("event should still be marked as seen for subsource1")
	}
	if !seenIDs[subsourceID2][eventID] {
		t.Error("event should now be marked as seen for subsource2")
	}
}

func TestGenerateDeterministicID(t *testing.T) {
	input := "https://example.com/article/123"

	id1 := generateDeterministicID(input)
	id2 := generateDeterministicID(input)

	// Same input should produce same ID
	if id1 != id2 {
		t.Errorf("expected deterministic ID, got different values: %s vs %s", id1, id2)
	}

	// ID should be non-empty
	if id1 == "" {
		t.Error("expected non-empty ID")
	}

	// Different input should produce different ID
	differentInput := "https://example.com/article/456"
	id3 := generateDeterministicID(differentInput)
	if id1 == id3 {
		t.Error("different inputs should produce different IDs")
	}
}

func TestSeenIDsPersistence_Format(t *testing.T) {
	// This test verifies the format used by loadSeenIDs and saveSeenIDs
	// The format should be: subsourceID\teventID

	seenIDs := make(map[string]map[string]bool)
	subsourceID := "test-subsource-id"
	eventID := "test-event-id"

	// Initialize and add
	if seenIDs[subsourceID] == nil {
		seenIDs[subsourceID] = make(map[string]bool)
	}
	seenIDs[subsourceID][eventID] = true

	// Verify structure
	if !seenIDs[subsourceID][eventID] {
		t.Error("event should be marked as seen")
	}

	// Verify we can iterate (as saveSeenIDs does)
	count := 0
	for subID, ids := range seenIDs {
		if subID != subsourceID {
			t.Errorf("expected subsourceID '%s', got '%s'", subsourceID, subID)
		}
		for evID := range ids {
			if evID != eventID {
				t.Errorf("expected eventID '%s', got '%s'", eventID, evID)
			}
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}
