package store

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: source-registration-ui, Property 1: Source Creation Round-Trip
// For any valid source configuration, when added to the store,
// it should be retrievable with all originally submitted fields intact.
func TestProperty_SourceCreationRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("created source can be retrieved with all fields", prop.ForAll(
		func(name, sourceType, webhook, secret string) bool {
			store := NewMemoryStore(100)

			// Create source with generated fields
			source := Source{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: webhook,
				WebhookSecret:  secret,
			}

			// Add to store
			err := store.AddSource(source)
			if err != nil {
				return false
			}

			// Retrieve all sources
			sources := store.ListSources()
			if len(sources) != 1 {
				return false
			}

			retrieved := sources[0]

			// Verify all fields match (except ID and CreatedAt which are generated)
			if retrieved.Name != name {
				return false
			}
			if retrieved.Type != sourceType {
				return false
			}
			if retrieved.DiscordWebhook != webhook {
				return false
			}
			if retrieved.WebhookSecret != secret {
				return false
			}
			if retrieved.ID == "" {
				return false
			}
			if retrieved.CreatedAt.IsZero() {
				return false
			}

			// Verify retrieval by ID works
			byID, found := store.GetSource(retrieved.ID)
			if !found {
				return false
			}
			if byID.Name != name || byID.Type != sourceType {
				return false
			}

			return true
		},
		genValidSourceName(),
		genValidSourceType(),
		genValidDiscordWebhook(),
		genOptionalSecret(),
	))

	properties.TestingRun(t)
}

// Generators for valid source fields

func genValidSourceName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 100
	})
}

func genValidSourceType() gopter.Gen {
	return gen.OneConstOf("github", "gitlab", "generic")
}

func genValidDiscordWebhook() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		return "https://discord.com/api/webhooks/" + s
	})
}

func genOptionalSecret() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) <= 256
	})
}

// Feature: source-registration-ui, Property 4: List Returns All Sources in Creation Order
// For any sequence of source creations, ListSources should return all created sources
// in the exact order they were created (oldest first).
func TestProperty_ListReturnsSourcesInCreationOrder(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("sources returned in creation order", prop.ForAll(
		func(count int) bool {
			if count < 2 || count > 10 {
				return true // Skip edge cases
			}

			store := NewMemoryStore(100)
			var createdSources []Source

			// Create multiple sources with slight time delays
			for i := 0; i < count; i++ {
				source := Source{
					Name:           "Source" + string(rune('A'+i)),
					Type:           "generic",
					DiscordWebhook: "https://discord.com/api/webhooks/test",
				}
				
				err := store.AddSource(source)
				if err != nil {
					return false
				}

				// Get the created source (with ID and timestamp)
				sources := store.ListSources()
				createdSources = append(createdSources, sources[len(sources)-1])

				// Small delay to ensure different timestamps
				time.Sleep(time.Millisecond)
			}

			// Retrieve all sources
			retrieved := store.ListSources()

			if len(retrieved) != count {
				return false
			}

			// Verify order matches creation order
			for i := 0; i < count; i++ {
				if retrieved[i].ID != createdSources[i].ID {
					return false
				}
				// Verify timestamps are in ascending order
				if i > 0 && retrieved[i].CreatedAt.Before(retrieved[i-1].CreatedAt) {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 10),
	))

	properties.TestingRun(t)
}

// Feature: source-registration-ui, Property 5: Duplicate Names Allowed with Unique IDs
// For any source name, creating multiple sources with the same name should succeed,
// and each should receive a distinct UUID identifier.
func TestProperty_DuplicateNamesAllowedWithUniqueIDs(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("duplicate names get unique IDs", prop.ForAll(
		func(name string, count int) bool {
			if count < 2 || count > 5 {
				return true // Test with 2-5 duplicates
			}

			store := NewMemoryStore(100)
			var ids []string

			// Create multiple sources with the same name
			for i := 0; i < count; i++ {
				source := Source{
					Name:           name,
					Type:           "generic",
					DiscordWebhook: "https://discord.com/api/webhooks/test",
				}

				err := store.AddSource(source)
				if err != nil {
					return false
				}

				// Get the created source
				sources := store.ListSources()
				if len(sources) != i+1 {
					return false
				}

				newSource := sources[len(sources)-1]
				ids = append(ids, newSource.ID)

				// Verify the name matches
				if newSource.Name != name {
					return false
				}
			}

			// Verify all IDs are unique
			idSet := make(map[string]bool)
			for _, id := range ids {
				if idSet[id] {
					return false // Duplicate ID found
				}
				idSet[id] = true
			}

			// Verify we have the expected number of unique IDs
			if len(idSet) != count {
				return false
			}

			return true
		},
		genValidSourceName(),
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t)
}
