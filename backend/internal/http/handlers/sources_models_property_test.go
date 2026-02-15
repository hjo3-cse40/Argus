package handlers

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: source-registration-ui, Property 2: Validation Rejects Invalid Inputs
// For any source configuration with invalid data (empty required fields, malformed Discord webhook URL),
// the validation should return an error.
func TestProperty_ValidationRejectsInvalidInputs(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Empty name should always fail validation
	properties.Property("empty name fails validation", prop.ForAll(
		func(sourceType, webhook string) bool {
			req := CreateSourceRequest{
				Name:           "",
				Type:           sourceType,
				DiscordWebhook: webhook,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceType(),
		genValidDiscordWebhook(),
	))

	// Property: Empty type should always fail validation
	properties.Property("empty type fails validation", prop.ForAll(
		func(name, webhook string) bool {
			req := CreateSourceRequest{
				Name:           name,
				Type:           "",
				DiscordWebhook: webhook,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceName(),
		genValidDiscordWebhook(),
	))

	// Property: Empty Discord webhook should always fail validation
	properties.Property("empty discord webhook fails validation", prop.ForAll(
		func(name, sourceType string) bool {
			req := CreateSourceRequest{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: "",
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceName(),
		genValidSourceType(),
	))

	// Property: Invalid Discord webhook URL should fail validation
	properties.Property("invalid discord webhook URL fails validation", prop.ForAll(
		func(name, sourceType, invalidWebhook string) bool {
			// Generate invalid webhooks (not starting with correct prefix)
			if len(invalidWebhook) < 4 || invalidWebhook[:4] != "http" {
				invalidWebhook = "http://example.com/" + invalidWebhook
			}

			req := CreateSourceRequest{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: invalidWebhook,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceName(),
		genValidSourceType(),
		gen.AlphaString(),
	))

	// Property: Invalid source type should fail validation
	properties.Property("invalid source type fails validation", prop.ForAll(
		func(name, webhook, invalidType string) bool {
			// Ensure it's not a valid type
			validTypes := map[string]bool{"github": true, "gitlab": true, "generic": true}
			if validTypes[invalidType] || invalidType == "" {
				invalidType = "invalid-" + invalidType
			}

			req := CreateSourceRequest{
				Name:           name,
				Type:           invalidType,
				DiscordWebhook: webhook,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceName(),
		genValidDiscordWebhook(),
		gen.AlphaString(),
	))

	// Property: Name exceeding max length should fail validation
	properties.Property("name exceeding 100 chars fails validation", prop.ForAll(
		func(sourceType, webhook string) bool {
			// Generate name longer than 100 characters
			longName := "a"
			for len(longName) <= 100 {
				longName += "a"
			}

			req := CreateSourceRequest{
				Name:           longName,
				Type:           sourceType,
				DiscordWebhook: webhook,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceType(),
		genValidDiscordWebhook(),
	))

	// Property: Webhook secret exceeding max length should fail validation
	properties.Property("webhook secret exceeding 256 chars fails validation", prop.ForAll(
		func(name, sourceType, webhook string) bool {
			// Generate secret longer than 256 characters
			longSecret := ""
			for len(longSecret) <= 256 {
				longSecret += "a"
			}

			req := CreateSourceRequest{
				Name:           name,
				Type:           sourceType,
				DiscordWebhook: webhook,
				WebhookSecret:  longSecret,
			}

			err := req.Validate()
			return err != nil // Should always have an error
		},
		genValidSourceName(),
		genValidSourceType(),
		genValidDiscordWebhook(),
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
		if s == "" {
			s = "test"
		}
		return "https://discord.com/api/webhooks/" + s
	})
}
