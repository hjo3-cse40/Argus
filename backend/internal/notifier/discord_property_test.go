package notifier

import (
	"fmt"
	"testing"

	"argus-backend/internal/events"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock store for testing
type mockStore struct {
	subsources map[string]Subsource
	platforms  map[string]Platform
	sources    map[string]Source
}

func newMockStore() *mockStore {
	return &mockStore{
		subsources: make(map[string]Subsource),
		platforms:  make(map[string]Platform),
		sources:    make(map[string]Source),
	}
}

func (m *mockStore) GetSubsource(id string) (Subsource, bool) {
	s, ok := m.subsources[id]
	return s, ok
}

func (m *mockStore) GetPlatform(id string) (Platform, bool) {
	p, ok := m.platforms[id]
	return p, ok
}

func (m *mockStore) GetSourceByName(name string) (Source, bool) {
	s, ok := m.sources[name]
	return s, ok
}

// Feature: hierarchical-sources, Property 20: Discord notification formatting with hierarchy
func TestProperty_DiscordNotificationFormattingWithHierarchy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: For any event with platform_name and subsource_name in metadata,
	// formatSource should return "{platform_name} - {subsource_name}"
	properties.Property("formatSource returns hierarchical format when metadata present",
		prop.ForAll(
			func(platformName, subsourceName, eventSource string) bool {
				event := events.NewEvent("test-id", eventSource, "Test Title", "https://example.com")
				event.Metadata["platform_name"] = platformName
				event.Metadata["subsource_name"] = subsourceName

				result := formatSource(event)
				expected := fmt.Sprintf("%s - %s", platformName, subsourceName)

				return result == expected
			},
			genPlatformName(),
			gen.Identifier(),
			gen.Identifier(),
		))

	// Property: For any event without platform_name or subsource_name,
	// formatSource should fall back to event.Source
	properties.Property("formatSource falls back to event.Source when metadata missing",
		prop.ForAll(
			func(eventSource string, missingField int) bool {
				event := events.NewEvent("test-id", eventSource, "Test Title", "https://example.com")

				// Randomly omit platform_name, subsource_name, or both
				switch missingField % 3 {
				case 0:
					// Missing platform_name
					event.Metadata["subsource_name"] = "subsource"
				case 1:
					// Missing subsource_name
					event.Metadata["platform_name"] = "youtube"
				case 2:
					// Missing both
				}

				result := formatSource(event)
				return result == eventSource
			},
			gen.Identifier(),
			gen.Int(),
		))

	// Property: For any event with subsource_id in metadata,
	// GetWebhookURL should retrieve webhook from platform via subsource
	properties.Property("GetWebhookURL retrieves webhook from platform via subsource",
		prop.ForAll(
			func(subsourceID, platformID, webhookURL string) bool {
				store := newMockStore()
				store.subsources[subsourceID] = Subsource{
					ID:         subsourceID,
					PlatformID: platformID,
					Name:       "test-subsource",
				}
				store.platforms[platformID] = Platform{
					ID:             platformID,
					Name:           "youtube",
					DiscordWebhook: webhookURL,
				}

				notifier := NewNotifier(store)
				event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
				event.Metadata["subsource_id"] = subsourceID

				result, err := notifier.GetWebhookURL(event)
				if err != nil {
					return false
				}

				return result == webhookURL
			},
			gen.Identifier(),
			gen.Identifier(),
			genDiscordWebhook(),
		))

	// Property: For any event without subsource_id,
	// GetWebhookURL should fall back to GetSourceByName
	properties.Property("GetWebhookURL falls back to GetSourceByName when subsource_id missing",
		prop.ForAll(
			func(sourceName, webhookURL string) bool {
				store := newMockStore()
				store.sources[sourceName] = Source{
					ID:             "source-id",
					Name:           sourceName,
					DiscordWebhook: webhookURL,
				}

				notifier := NewNotifier(store)
				event := events.NewEvent("test-id", sourceName, "Test Title", "https://example.com")
				// No subsource_id in metadata

				result, err := notifier.GetWebhookURL(event)
				if err != nil {
					return false
				}

				return result == webhookURL
			},
			gen.Identifier(),
			genDiscordWebhook(),
		))

	// Property: GetWebhookURL returns error when subsource not found
	properties.Property("GetWebhookURL returns error when subsource not found",
		prop.ForAll(
			func(subsourceID string) bool {
				store := newMockStore()
				notifier := NewNotifier(store)
				event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
				event.Metadata["subsource_id"] = subsourceID

				_, err := notifier.GetWebhookURL(event)
				return err != nil
			},
			gen.Identifier(),
		))

	// Property: GetWebhookURL returns error when platform not found
	properties.Property("GetWebhookURL returns error when platform not found",
		prop.ForAll(
			func(subsourceID, platformID string) bool {
				store := newMockStore()
				store.subsources[subsourceID] = Subsource{
					ID:         subsourceID,
					PlatformID: platformID,
					Name:       "test-subsource",
				}
				// Platform not in store

				notifier := NewNotifier(store)
				event := events.NewEvent("test-id", "test-source", "Test Title", "https://example.com")
				event.Metadata["subsource_id"] = subsourceID

				_, err := notifier.GetWebhookURL(event)
				return err != nil
			},
			gen.Identifier(),
			gen.Identifier(),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Generators

func genPlatformName() gopter.Gen {
	return gen.OneConstOf("youtube", "reddit", "x")
}

func genDiscordWebhook() gopter.Gen {
	return gen.Identifier().Map(func(id string) string {
		return fmt.Sprintf("https://discord.com/api/webhooks/%s", id)
	})
}
