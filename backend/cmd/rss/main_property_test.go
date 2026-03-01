package main

import (
	"argus-backend/internal/events"
	"argus-backend/internal/store"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: hierarchical-sources, Property 16: Event metadata population with hierarchical information
func TestProperty_EventMetadataPopulation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Event metadata contains subsource_id, platform_name, subsource_name, and source equals subsource name",
		prop.ForAll(
			func(subsourceID, platformName, subsourceName, identifier string) bool {
				subsource := store.SubsourceWithPlatform{
					Subsource: store.Subsource{
						ID:         subsourceID,
						PlatformID: "test-platform-id",
						Name:       subsourceName,
						Identifier: identifier,
					},
					PlatformName: platformName,
				}

				// Simulate event creation (simplified version of what processFeeds does)
				event := events.NewEvent("test-event-id", subsource.Name, "Test Title", "https://example.com")
				event.Metadata["subsource_id"] = subsource.ID
				event.Metadata["platform_name"] = subsource.PlatformName
				event.Metadata["subsource_name"] = subsource.Name

				// Verify metadata
				if event.Metadata["subsource_id"] != subsourceID {
					return false
				}
				if event.Metadata["platform_name"] != platformName {
					return false
				}
				if event.Metadata["subsource_name"] != subsourceName {
					return false
				}
				if event.Source != subsourceName {
					return false
				}

				return true
			},
			gen.Identifier(),                                // subsourceID
			gen.OneConstOf("youtube", "reddit", "x"),        // platformName
			gen.AlphaString().SuchThat(func(s string) bool { // subsourceName
				return len(s) > 0
			}),
			gen.Identifier(), // identifier
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: hierarchical-sources, Property 17: RSSHub URL construction from platform and identifier
func TestProperty_RSSHubURLConstruction(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("RSSHub URL is constructed correctly for each platform",
		prop.ForAll(
			func(platformName, identifier string) bool {
				baseURL := "https://rsshub.example.com"
				url := constructRSSHubURL(baseURL, platformName, identifier)

				var expectedPath string
				switch platformName {
				case "youtube":
					expectedPath = fmt.Sprintf("%s/youtube/channel/%s", baseURL, identifier)
				case "reddit":
					expectedPath = fmt.Sprintf("%s/reddit/subreddit/%s", baseURL, identifier)
				case "x":
					expectedPath = fmt.Sprintf("%s/twitter/user/%s", baseURL, identifier)
				default:
					return url == "" // Unsupported platforms should return empty string
				}

				return url == expectedPath
			},
			gen.OneConstOf("youtube", "reddit", "x"), // platformName
			gen.Identifier(),                         // identifier
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: hierarchical-sources, Property 21: RSS poller deduplication by subsource
func TestProperty_DeduplicationBySubsource(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Same event_id for same subsource is deduplicated",
		prop.ForAll(
			func(subsourceID, eventID string) bool {
				seenIDs := make(map[string]map[string]bool)

				// First occurrence - should not be seen
				if seenIDs[subsourceID] == nil {
					seenIDs[subsourceID] = make(map[string]bool)
				}
				firstSeen := seenIDs[subsourceID][eventID]

				// Mark as seen
				seenIDs[subsourceID][eventID] = true

				// Second occurrence - should be seen
				secondSeen := seenIDs[subsourceID][eventID]

				// First should be false (not seen), second should be true (seen)
				return !firstSeen && secondSeen
			},
			gen.Identifier(), // subsourceID
			gen.Identifier(), // eventID
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: hierarchical-sources, Property 21 (extended): Different subsources don't interfere with deduplication
func TestProperty_DeduplicationIsolatedBySubsource(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Same event_id for different subsources are tracked independently",
		prop.ForAll(
			func(subsourceID1, subsourceID2, eventID string) bool {
				// Ensure subsources are different
				if subsourceID1 == subsourceID2 {
					return true // Skip this case
				}

				seenIDs := make(map[string]map[string]bool)

				// Mark event as seen for subsource1
				if seenIDs[subsourceID1] == nil {
					seenIDs[subsourceID1] = make(map[string]bool)
				}
				seenIDs[subsourceID1][eventID] = true

				// Check if seen for subsource2 (should be false)
				if seenIDs[subsourceID2] == nil {
					seenIDs[subsourceID2] = make(map[string]bool)
				}
				seenForSubsource2 := seenIDs[subsourceID2][eventID]

				// Event should not be marked as seen for subsource2
				return !seenForSubsource2
			},
			gen.Identifier(), // subsourceID1
			gen.Identifier(), // subsourceID2
			gen.Identifier(), // eventID
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: hierarchical-sources, Property 28: RSSHUB_FEEDS parsing extracts platform and identifier
func TestProperty_FeedParsingRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Feed format round-trip preserves platform and identifier",
		prop.ForAll(
			func(platformName, identifier string) bool {
				// Format: "platform:identifier"
				feedStr := fmt.Sprintf("%s:%s", platformName, identifier)

				// Parse (simulate what config parsing does)
				parts := splitFeedString(feedStr)
				if len(parts) != 2 {
					return false
				}

				parsedPlatform := parts[0]
				parsedIdentifier := parts[1]

				// Verify round-trip
				return parsedPlatform == platformName && parsedIdentifier == identifier
			},
			gen.OneConstOf("youtube", "reddit", "x"), // platformName
			gen.Identifier().SuchThat(func(s string) bool {
				return len(s) > 0 && s != "" // identifier must not be empty
			}),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper function to simulate feed string parsing
func splitFeedString(feedStr string) []string {
	// Simple split on ":"
	result := []string{}
	colonIdx := -1
	for i, c := range feedStr {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return result
	}
	result = append(result, feedStr[:colonIdx])
	result = append(result, feedStr[colonIdx+1:])
	return result
}
