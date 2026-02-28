package notifier

import (
	"strings"
	"testing"

	"argus-backend/internal/events"
)

// Test formatSource with platform_name and subsource_name returns "Platform - Subsource"
func TestFormatSource_WithHierarchicalMetadata(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["platform_name"] = "youtube"
	event.Metadata["subsource_name"] = "NBA"

	result := formatSource(event)
	expected := "youtube - NBA"

	if result != expected {
		t.Errorf("formatSource() = %q, want %q", result, expected)
	}
}

// Test formatSource with missing platform_name falls back to event.Source
func TestFormatSource_MissingPlatformName(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["subsource_name"] = "NBA"
	// platform_name is missing

	result := formatSource(event)
	expected := "fallback-source"

	if result != expected {
		t.Errorf("formatSource() = %q, want %q", result, expected)
	}
}

// Test formatSource with missing subsource_name falls back to event.Source
func TestFormatSource_MissingSubsourceName(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["platform_name"] = "youtube"
	// subsource_name is missing

	result := formatSource(event)
	expected := "fallback-source"

	if result != expected {
		t.Errorf("formatSource() = %q, want %q", result, expected)
	}
}

// Test formatSource with empty platform_name falls back to event.Source
func TestFormatSource_EmptyPlatformName(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["platform_name"] = ""
	event.Metadata["subsource_name"] = "NBA"

	result := formatSource(event)
	expected := "fallback-source"

	if result != expected {
		t.Errorf("formatSource() = %q, want %q", result, expected)
	}
}

// Test formatSource with empty subsource_name falls back to event.Source
func TestFormatSource_EmptySubsourceName(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["platform_name"] = "youtube"
	event.Metadata["subsource_name"] = ""

	result := formatSource(event)
	expected := "fallback-source"

	if result != expected {
		t.Errorf("formatSource() = %q, want %q", result, expected)
	}
}

// Test getWebhookURL with subsource_id retrieves webhook from platform
func TestGetWebhookURL_WithSubsourceID(t *testing.T) {
	store := newMockStore()
	store.subsources["sub-123"] = Subsource{
		ID:         "sub-123",
		PlatformID: "plat-456",
		Name:       "NBA",
	}
	store.platforms["plat-456"] = Platform{
		ID:             "plat-456",
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/test123",
	}

	notifier := NewNotifier(store)
	event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
	event.Metadata["subsource_id"] = "sub-123"

	result, err := notifier.GetWebhookURL(event)
	if err != nil {
		t.Fatalf("GetWebhookURL() error = %v, want nil", err)
	}

	expected := "https://discord.com/api/webhooks/test123"
	if result != expected {
		t.Errorf("GetWebhookURL() = %q, want %q", result, expected)
	}
}

// Test getWebhookURL with missing subsource_id falls back to GetSourceByName
func TestGetWebhookURL_FallbackToSourceByName(t *testing.T) {
	store := newMockStore()
	store.sources["legacy-source"] = Source{
		ID:             "src-789",
		Name:           "legacy-source",
		DiscordWebhook: "https://discord.com/api/webhooks/legacy123",
	}

	notifier := NewNotifier(store)
	event := events.NewEvent("test-id", "legacy-source", "Test Title", "https://example.com")
	// No subsource_id in metadata

	result, err := notifier.GetWebhookURL(event)
	if err != nil {
		t.Fatalf("GetWebhookURL() error = %v, want nil", err)
	}

	expected := "https://discord.com/api/webhooks/legacy123"
	if result != expected {
		t.Errorf("GetWebhookURL() = %q, want %q", result, expected)
	}
}

// Test getWebhookURL logs error when subsource not found
func TestGetWebhookURL_SubsourceNotFound(t *testing.T) {
	store := newMockStore()
	notifier := NewNotifier(store)
	event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
	event.Metadata["subsource_id"] = "non-existent-sub"

	_, err := notifier.GetWebhookURL(event)
	if err == nil {
		t.Fatal("GetWebhookURL() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "subsource not found") {
		t.Errorf("GetWebhookURL() error = %q, want error containing 'subsource not found'", err.Error())
	}
}

// Test getWebhookURL logs error when platform not found
func TestGetWebhookURL_PlatformNotFound(t *testing.T) {
	store := newMockStore()
	store.subsources["sub-123"] = Subsource{
		ID:         "sub-123",
		PlatformID: "non-existent-plat",
		Name:       "NBA",
	}

	notifier := NewNotifier(store)
	event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
	event.Metadata["subsource_id"] = "sub-123"

	_, err := notifier.GetWebhookURL(event)
	if err == nil {
		t.Fatal("GetWebhookURL() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "platform not found") {
		t.Errorf("GetWebhookURL() error = %q, want error containing 'platform not found'", err.Error())
	}
}

// Test getWebhookURL with empty subsource_id falls back to GetSourceByName
func TestGetWebhookURL_EmptySubsourceID(t *testing.T) {
	store := newMockStore()
	store.sources["legacy-source"] = Source{
		ID:             "src-789",
		Name:           "legacy-source",
		DiscordWebhook: "https://discord.com/api/webhooks/legacy123",
	}

	notifier := NewNotifier(store)
	event := events.NewEvent("test-id", "legacy-source", "Test Title", "https://example.com")
	event.Metadata["subsource_id"] = "" // Empty string

	result, err := notifier.GetWebhookURL(event)
	if err != nil {
		t.Fatalf("GetWebhookURL() error = %v, want nil", err)
	}

	expected := "https://discord.com/api/webhooks/legacy123"
	if result != expected {
		t.Errorf("GetWebhookURL() = %q, want %q", result, expected)
	}
}

// Test embed rendering uses formatSource for source display
func TestRenderDiscordEmbed_UsesFormatSource(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	event.Metadata["platform_name"] = "youtube"
	event.Metadata["subsource_name"] = "NBA"
	event.Metadata["author"] = "Test Author"
	event.Metadata["description"] = "Test description"

	payload := RenderDiscordEmbed(event)

	if len(payload.Embeds) != 1 {
		t.Fatalf("RenderDiscordEmbed() returned %d embeds, want 1", len(payload.Embeds))
	}

	embed := payload.Embeds[0]

	// Check that the description contains the formatted source
	expectedSource := "**Source:** youtube - NBA"
	if !strings.Contains(embed.Description, expectedSource) {
		t.Errorf("RenderDiscordEmbed() description = %q, want to contain %q", embed.Description, expectedSource)
	}

	// Verify other fields are maintained
	if embed.Title != "Test Title" {
		t.Errorf("RenderDiscordEmbed() title = %q, want %q", embed.Title, "Test Title")
	}

	if embed.URL != "https://example.com" {
		t.Errorf("RenderDiscordEmbed() url = %q, want %q", embed.URL, "https://example.com")
	}

	if !strings.Contains(embed.Description, "**Author:** Test Author") {
		t.Errorf("RenderDiscordEmbed() description should contain author")
	}

	if embed.Footer == nil || !strings.Contains(embed.Footer.Text, "event_id: test-id") {
		t.Errorf("RenderDiscordEmbed() footer should contain event_id")
	}

	if embed.Timestamp == "" {
		t.Errorf("RenderDiscordEmbed() timestamp should not be empty")
	}
}

// Test embed rendering falls back to event.Source when metadata missing
func TestRenderDiscordEmbed_FallbackToEventSource(t *testing.T) {
	event := events.NewEvent("test-id", "fallback-source", "Test Title", "https://example.com")
	// No hierarchical metadata

	payload := RenderDiscordEmbed(event)

	if len(payload.Embeds) != 1 {
		t.Fatalf("RenderDiscordEmbed() returned %d embeds, want 1", len(payload.Embeds))
	}

	embed := payload.Embeds[0]

	// Check that the description contains the fallback source
	expectedSource := "**Source:** fallback-source"
	if !strings.Contains(embed.Description, expectedSource) {
		t.Errorf("RenderDiscordEmbed() description = %q, want to contain %q", embed.Description, expectedSource)
	}
}
