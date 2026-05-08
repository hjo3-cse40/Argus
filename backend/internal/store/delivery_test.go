package store

import (
	"testing"
	"time"
)

// Test AddQueued populates subsource_id from event metadata
func TestAddQueued_PopulatesSubsourceID(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)
	subsourceID := "sub-123"

	delivery := Delivery{
		EventID:     "event-1",
		Source:      "test-source",
		Title:       "Test Title",
		URL:         "https://example.com",
		SubsourceID: &subsourceID,
		UserID:      owner,
	}

	store.AddQueued(delivery)

	deliveries := store.List(owner)
	if len(deliveries) != 1 {
		t.Fatalf("Expected 1 delivery, got %d", len(deliveries))
	}

	if deliveries[0].SubsourceID == nil {
		t.Fatal("Expected subsource_id to be populated, got nil")
	}

	if *deliveries[0].SubsourceID != subsourceID {
		t.Errorf("Expected subsource_id %q, got %q", subsourceID, *deliveries[0].SubsourceID)
	}
}

// Test AddQueued with nil subsource_id
func TestAddQueued_NilSubsourceID(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)

	delivery := Delivery{
		EventID:     "event-1",
		Source:      "test-source",
		Title:       "Test Title",
		URL:         "https://example.com",
		SubsourceID: nil,
		UserID:      owner,
	}

	store.AddQueued(delivery)

	deliveries := store.List(owner)
	if len(deliveries) != 1 {
		t.Fatalf("Expected 1 delivery, got %d", len(deliveries))
	}

	if deliveries[0].SubsourceID != nil {
		t.Errorf("Expected subsource_id to be nil, got %q", *deliveries[0].SubsourceID)
	}
}

// Test ListDeliveriesBySubsource returns only deliveries for that subsource
func TestListDeliveriesBySubsource_FiltersCorrectly(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)
	subsource1 := "sub-1"
	subsource2 := "sub-2"

	// Add deliveries for subsource 1
	for i := 0; i < 3; i++ {
		delivery := Delivery{
			EventID:     "event-sub1-" + string(rune(i)),
			Source:      "test-source",
			Title:       "Test Title",
			URL:         "https://example.com",
			SubsourceID: &subsource1,
			UserID:      owner,
		}
		store.AddQueued(delivery)
	}

	// Add deliveries for subsource 2
	for i := 0; i < 2; i++ {
		delivery := Delivery{
			EventID:     "event-sub2-" + string(rune(i)),
			Source:      "test-source",
			Title:       "Test Title",
			URL:         "https://example.com",
			SubsourceID: &subsource2,
			UserID:      owner,
		}
		store.AddQueued(delivery)
	}

	filtered := store.ListDeliveriesBySubsource(owner, subsource1)

	if len(filtered) != 3 {
		t.Errorf("Expected 3 deliveries for subsource 1, got %d", len(filtered))
	}

	for _, d := range filtered {
		if d.SubsourceID == nil || *d.SubsourceID != subsource1 {
			t.Errorf("Expected all deliveries to have subsource_id %q", subsource1)
		}
	}
}

// Test ListDeliveriesBySubsource with non-existent subsource returns empty
func TestListDeliveriesBySubsource_NonExistentReturnsEmpty(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)
	subsource1 := "sub-1"

	delivery := Delivery{
		EventID:     "event-1",
		Source:      "test-source",
		Title:       "Test Title",
		URL:         "https://example.com",
		SubsourceID: &subsource1,
		UserID:      owner,
	}
	store.AddQueued(delivery)

	filtered := store.ListDeliveriesBySubsource(owner, "non-existent")

	if len(filtered) != 0 {
		t.Errorf("Expected 0 deliveries for non-existent subsource, got %d", len(filtered))
	}
}

// Test ListDeliveriesByPlatform returns only deliveries for subsources of that platform
func TestListDeliveriesByPlatform_FiltersCorrectly(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)

	platform1 := Platform{
		ID:             "plat-1",
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/1",
		CreatedAt:      time.Now().UTC(),
	}
	platform2 := Platform{
		ID:             "plat-2",
		Name:           "reddit",
		DiscordWebhook: "https://discord.com/api/webhooks/2",
		CreatedAt:      time.Now().UTC(),
	}

	err := store.AddPlatform(owner, platform1)
	if err != nil {
		t.Fatalf("Failed to add platform 1: %v", err)
	}
	err = store.AddPlatform(owner, platform2)
	if err != nil {
		t.Fatalf("Failed to add platform 2: %v", err)
	}

	subsource1 := Subsource{
		ID:         "sub-1",
		PlatformID: "plat-1",
		Name:       "Subsource 1",
		Identifier: "id-1",
		CreatedAt:  time.Now().UTC(),
	}
	subsource2 := Subsource{
		ID:         "sub-2",
		PlatformID: "plat-2",
		Name:       "Subsource 2",
		Identifier: "id-2",
		CreatedAt:  time.Now().UTC(),
	}

	err = store.AddSubsource(owner, subsource1)
	if err != nil {
		t.Fatalf("Failed to add subsource 1: %v", err)
	}
	err = store.AddSubsource(owner, subsource2)
	if err != nil {
		t.Fatalf("Failed to add subsource 2: %v", err)
	}

	sub1ID := "sub-1"
	for i := 0; i < 3; i++ {
		delivery := Delivery{
			EventID:     "event-plat1-" + string(rune(i)),
			Source:      "test-source",
			Title:       "Test Title",
			URL:         "https://example.com",
			SubsourceID: &sub1ID,
			UserID:      owner,
		}
		store.AddQueued(delivery)
	}

	sub2ID := "sub-2"
	for i := 0; i < 2; i++ {
		delivery := Delivery{
			EventID:     "event-plat2-" + string(rune(i)),
			Source:      "test-source",
			Title:       "Test Title",
			URL:         "https://example.com",
			SubsourceID: &sub2ID,
			UserID:      owner,
		}
		store.AddQueued(delivery)
	}

	filtered := store.ListDeliveriesByPlatform(owner, "plat-1")

	if len(filtered) != 3 {
		t.Errorf("Expected 3 deliveries for platform 1, got %d", len(filtered))
	}

	for _, d := range filtered {
		if d.SubsourceID == nil || *d.SubsourceID != "sub-1" {
			t.Errorf("Expected all deliveries to have subsource_id from platform 1")
		}
	}
}

// Test ListDeliveriesByPlatform with non-existent platform returns empty
func TestListDeliveriesByPlatform_NonExistentReturnsEmpty(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)

	platform := Platform{
		ID:             "plat-1",
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/1",
		CreatedAt:      time.Now().UTC(),
	}
	err := store.AddPlatform(owner, platform)
	if err != nil {
		t.Fatalf("Failed to add platform: %v", err)
	}

	subsource := Subsource{
		ID:         "sub-1",
		PlatformID: "plat-1",
		Name:       "Subsource 1",
		Identifier: "id-1",
		CreatedAt:  time.Now().UTC(),
	}
	err = store.AddSubsource(owner, subsource)
	if err != nil {
		t.Fatalf("Failed to add subsource: %v", err)
	}

	subID := "sub-1"
	delivery := Delivery{
		EventID:     "event-1",
		Source:      "test-source",
		Title:       "Test Title",
		URL:         "https://example.com",
		SubsourceID: &subID,
		UserID:      owner,
	}
	store.AddQueued(delivery)

	filtered := store.ListDeliveriesByPlatform(owner, "non-existent")

	if len(filtered) != 0 {
		t.Errorf("Expected 0 deliveries for non-existent platform, got %d", len(filtered))
	}
}

// Test List includes subsource_id in results
func TestList_IncludesSubsourceID(t *testing.T) {
	store := NewMemoryStore(100)
	owner := memoryTestUser(t, store)
	subsourceID := "sub-123"

	delivery := Delivery{
		EventID:     "event-1",
		Source:      "test-source",
		Title:       "Test Title",
		URL:         "https://example.com",
		SubsourceID: &subsourceID,
		UserID:      owner,
	}

	store.AddQueued(delivery)

	deliveries := store.List(owner)
	if len(deliveries) != 1 {
		t.Fatalf("Expected 1 delivery, got %d", len(deliveries))
	}

	if deliveries[0].SubsourceID == nil {
		t.Fatal("Expected subsource_id in list results, got nil")
	}

	if *deliveries[0].SubsourceID != subsourceID {
		t.Errorf("Expected subsource_id %q in list results, got %q", subsourceID, *deliveries[0].SubsourceID)
	}
}
