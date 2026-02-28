package store

import (
	"strings"
	"testing"
)

func TestValidatePlatform_ValidNames(t *testing.T) {
	validNames := []string{"youtube", "reddit", "x"}
	
	for _, name := range validNames {
		t.Run("valid_name_"+name, func(t *testing.T) {
			platform := Platform{
				Name:           name,
				DiscordWebhook: "https://discord.com/api/webhooks/123456789/abcdef",
			}
			
			err := validatePlatform(platform)
			if err != nil {
				t.Errorf("validatePlatform() should accept valid name %q, got error: %v", name, err)
			}
		})
	}
}

func TestValidatePlatform_InvalidNames(t *testing.T) {
	invalidNames := []string{"tiktok", "facebook", "empty", "", "   ", "YOUTUBE", "Reddit", "X"}
	
	for _, name := range invalidNames {
		t.Run("invalid_name_"+name, func(t *testing.T) {
			platform := Platform{
				Name:           name,
				DiscordWebhook: "https://discord.com/api/webhooks/123456789/abcdef",
			}
			
			err := validatePlatform(platform)
			if err == nil {
				t.Errorf("validatePlatform() should reject invalid name %q", name)
			}
			
			// Check that the error message mentions the name validation
			if !strings.Contains(err.Error(), "name") {
				t.Errorf("validatePlatform() error should mention name validation, got: %v", err)
			}
		})
	}
}

func TestValidatePlatform_ValidWebhookURLs(t *testing.T) {
	validWebhooks := []string{
		"https://discord.com/api/webhooks/123456789/abcdef",
		"https://discord.com/api/webhooks/987654321/xyz123",
		"https://discord.com/api/webhooks/111/aaa",
	}
	
	for _, webhook := range validWebhooks {
		t.Run("valid_webhook", func(t *testing.T) {
			platform := Platform{
				Name:           "youtube",
				DiscordWebhook: webhook,
			}
			
			err := validatePlatform(platform)
			if err != nil {
				t.Errorf("validatePlatform() should accept valid webhook %q, got error: %v", webhook, err)
			}
		})
	}
}

func TestValidatePlatform_InvalidWebhookURLs(t *testing.T) {
	invalidWebhooks := []string{
		"https://example.com/webhook",
		"http://discord.com/api/webhooks/123",
		"discord.com/api/webhooks/123",
		"https://discord.com/webhooks/123",
		"",
		"   ",
		"not-a-url",
	}
	
	for _, webhook := range invalidWebhooks {
		t.Run("invalid_webhook", func(t *testing.T) {
			platform := Platform{
				Name:           "youtube",
				DiscordWebhook: webhook,
			}
			
			err := validatePlatform(platform)
			if err == nil {
				t.Errorf("validatePlatform() should reject invalid webhook %q", webhook)
			}
			
			// Check that the error message mentions the webhook validation
			if !strings.Contains(err.Error(), "discord_webhook") {
				t.Errorf("validatePlatform() error should mention webhook validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_ValidData(t *testing.T) {
	validSubsources := []Subsource{
		{
			Name:       "NBA",
			Identifier: "UCxxx",
			URL:        "https://youtube.com/channel/UCxxx",
		},
		{
			Name:       "r/programming",
			Identifier: "programming",
			URL:        "https://reddit.com/r/programming",
		},
		{
			Name:       "elonmusk",
			Identifier: "elonmusk",
			URL:        "https://x.com/elonmusk",
		},
		{
			Name:       "Test Channel",
			Identifier: "test123",
			URL:        "", // Empty URL should be allowed
		},
	}
	
	for i, subsource := range validSubsources {
		t.Run("valid_subsource_"+string(rune('0'+i)), func(t *testing.T) {
			err := validateSubsource(subsource)
			if err != nil {
				t.Errorf("validateSubsource() should accept valid subsource %+v, got error: %v", subsource, err)
			}
		})
	}
}

func TestValidateSubsource_EmptyName(t *testing.T) {
	emptyNames := []string{"", "   ", "\t", "\n", "  \t  "}
	
	for _, name := range emptyNames {
		t.Run("empty_name", func(t *testing.T) {
			subsource := Subsource{
				Name:       name,
				Identifier: "valid_identifier",
			}
			
			err := validateSubsource(subsource)
			if err == nil {
				t.Errorf("validateSubsource() should reject empty/whitespace name %q", name)
			}
			
			// Check that the error message mentions the name validation
			if !strings.Contains(err.Error(), "name cannot be empty") {
				t.Errorf("validateSubsource() error should mention name validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_WhitespaceOnlyName(t *testing.T) {
	whitespaceNames := []string{"   ", "\t", "\n", "  \t  ", "\n\t\n"}
	
	for _, name := range whitespaceNames {
		t.Run("whitespace_name", func(t *testing.T) {
			subsource := Subsource{
				Name:       name,
				Identifier: "valid_identifier",
			}
			
			err := validateSubsource(subsource)
			if err == nil {
				t.Errorf("validateSubsource() should reject whitespace-only name %q", name)
			}
			
			// Check that the error message mentions the name validation
			if !strings.Contains(err.Error(), "name cannot be empty") {
				t.Errorf("validateSubsource() error should mention name validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_EmptyIdentifier(t *testing.T) {
	emptyIdentifiers := []string{"", "   ", "\t", "\n", "  \t  "}
	
	for _, identifier := range emptyIdentifiers {
		t.Run("empty_identifier", func(t *testing.T) {
			subsource := Subsource{
				Name:       "Valid Name",
				Identifier: identifier,
			}
			
			err := validateSubsource(subsource)
			if err == nil {
				t.Errorf("validateSubsource() should reject empty/whitespace identifier %q", identifier)
			}
			
			// Check that the error message mentions the identifier validation
			if !strings.Contains(err.Error(), "identifier cannot be empty") {
				t.Errorf("validateSubsource() error should mention identifier validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_WhitespaceOnlyIdentifier(t *testing.T) {
	whitespaceIdentifiers := []string{"   ", "\t", "\n", "  \t  ", "\n\t\n"}
	
	for _, identifier := range whitespaceIdentifiers {
		t.Run("whitespace_identifier", func(t *testing.T) {
			subsource := Subsource{
				Name:       "Valid Name",
				Identifier: identifier,
			}
			
			err := validateSubsource(subsource)
			if err == nil {
				t.Errorf("validateSubsource() should reject whitespace-only identifier %q", identifier)
			}
			
			// Check that the error message mentions the identifier validation
			if !strings.Contains(err.Error(), "identifier cannot be empty") {
				t.Errorf("validateSubsource() error should mention identifier validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_MalformedURLs(t *testing.T) {
	malformedURLs := []string{
		"not-a-url",
		"://missing-scheme",
		"relative/path",
		"just-text",
	}
	
	for _, url := range malformedURLs {
		t.Run("malformed_url", func(t *testing.T) {
			subsource := Subsource{
				Name:       "Valid Name",
				Identifier: "valid_identifier",
				URL:        url,
			}
			
			err := validateSubsource(subsource)
			if err == nil {
				t.Errorf("validateSubsource() should reject malformed URL %q", url)
			}
			
			// Check that the error message mentions the URL validation
			if !strings.Contains(err.Error(), "url must") {
				t.Errorf("validateSubsource() error should mention URL validation, got: %v", err)
			}
		})
	}
}

func TestValidateSubsource_ValidURLs(t *testing.T) {
	validURLs := []string{
		"https://youtube.com/channel/UCxxx",
		"https://reddit.com/r/programming",
		"https://x.com/elonmusk",
		"http://example.com",
		"https://example.com/path?query=value",
		"", // Empty URL should be allowed
	}
	
	for _, url := range validURLs {
		t.Run("valid_url", func(t *testing.T) {
			subsource := Subsource{
				Name:       "Valid Name",
				Identifier: "valid_identifier",
				URL:        url,
			}
			
			err := validateSubsource(subsource)
			if err != nil {
				t.Errorf("validateSubsource() should accept valid URL %q, got error: %v", url, err)
			}
		})
	}
}

// Integration tests with store operations

func TestMemoryStore_AddPlatform_ValidationIntegration(t *testing.T) {
	store := NewMemoryStore(100)
	
	// Test valid platform
	validPlatform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123456789/abcdef",
	}
	
	err := store.AddPlatform(validPlatform)
	if err != nil {
		t.Errorf("AddPlatform() should accept valid platform, got error: %v", err)
	}
	
	// Test invalid platform name
	invalidPlatform := Platform{
		Name:           "tiktok",
		DiscordWebhook: "https://discord.com/api/webhooks/123456789/abcdef",
	}
	
	err = store.AddPlatform(invalidPlatform)
	if err == nil {
		t.Error("AddPlatform() should reject invalid platform name")
	}
	
	// Test invalid webhook
	invalidWebhookPlatform := Platform{
		Name:           "reddit",
		DiscordWebhook: "https://example.com/webhook",
	}
	
	err = store.AddPlatform(invalidWebhookPlatform)
	if err == nil {
		t.Error("AddPlatform() should reject invalid webhook URL")
	}
}

func TestMemoryStore_AddSubsource_ValidationIntegration(t *testing.T) {
	store := NewMemoryStore(100)
	
	// First create a platform
	platform := Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123456789/abcdef",
	}
	err := store.AddPlatform(platform)
	if err != nil {
		t.Fatalf("Failed to create platform: %v", err)
	}
	
	platforms := store.ListPlatforms()
	if len(platforms) == 0 {
		t.Fatal("No platforms found")
	}
	platformID := platforms[0].ID
	
	// Test valid subsource
	validSubsource := Subsource{
		PlatformID: platformID,
		Name:       "NBA",
		Identifier: "UCxxx",
	}
	
	err = store.AddSubsource(validSubsource)
	if err != nil {
		t.Errorf("AddSubsource() should accept valid subsource, got error: %v", err)
	}
	
	// Test empty name
	emptyNameSubsource := Subsource{
		PlatformID: platformID,
		Name:       "",
		Identifier: "UCyyy",
	}
	
	err = store.AddSubsource(emptyNameSubsource)
	if err == nil {
		t.Error("AddSubsource() should reject empty name")
	}
	
	// Test empty identifier
	emptyIdentifierSubsource := Subsource{
		PlatformID: platformID,
		Name:       "NFL",
		Identifier: "",
	}
	
	err = store.AddSubsource(emptyIdentifierSubsource)
	if err == nil {
		t.Error("AddSubsource() should reject empty identifier")
	}
	
	// Test malformed URL
	malformedURLSubsource := Subsource{
		PlatformID: platformID,
		Name:       "MLB",
		Identifier: "UCzzz",
		URL:        "not-a-url",
	}
	
	err = store.AddSubsource(malformedURLSubsource)
	if err == nil {
		t.Error("AddSubsource() should reject malformed URL")
	}
}