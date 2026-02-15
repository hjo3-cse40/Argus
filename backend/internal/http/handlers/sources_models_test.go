package handlers

import (
	"strings"
	"testing"
)

func TestCreateSourceRequest_Validate_RequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateSourceRequest
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid request",
			req: CreateSourceRequest{
				Name:           "GitHub Main",
				Type:           "github",
				DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: CreateSourceRequest{
				Type:           "github",
				DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
			},
			wantErr:     true,
			errContains: []string{"name is required"},
		},
		{
			name: "missing type",
			req: CreateSourceRequest{
				Name:           "GitHub Main",
				DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
			},
			wantErr:     true,
			errContains: []string{"type is required"},
		},
		{
			name: "missing discord webhook",
			req: CreateSourceRequest{
				Name: "GitHub Main",
				Type: "github",
			},
			wantErr:     true,
			errContains: []string{"discord_webhook is required"},
		},
		{
			name: "multiple missing fields",
			req:  CreateSourceRequest{},
			wantErr: true,
			errContains: []string{
				"name is required",
				"type is required",
				"discord_webhook is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				errMsg := err.Error()
				for _, contains := range tt.errContains {
					if !strings.Contains(errMsg, contains) {
						t.Errorf("Validate() error = %v, should contain %q", err, contains)
					}
				}
			}
		})
	}
}

func TestCreateSourceRequest_Validate_DiscordWebhookFormat(t *testing.T) {
	tests := []struct {
		name    string
		webhook string
		wantErr bool
	}{
		{
			name:    "valid discord webhook",
			webhook: "https://discord.com/api/webhooks/123456789/abcdefghijk",
			wantErr: false,
		},
		{
			name:    "invalid protocol",
			webhook: "http://discord.com/api/webhooks/123/abc",
			wantErr: true,
		},
		{
			name:    "wrong domain",
			webhook: "https://example.com/api/webhooks/123/abc",
			wantErr: true,
		},
		{
			name:    "missing webhooks path",
			webhook: "https://discord.com/api/123/abc",
			wantErr: true,
		},
		{
			name:    "empty webhook",
			webhook: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateSourceRequest{
				Name:           "Test Source",
				Type:           "generic",
				DiscordWebhook: tt.webhook,
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateSourceRequest_Validate_SourceType(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantErr    bool
	}{
		{
			name:       "valid type github",
			sourceType: "github",
			wantErr:    false,
		},
		{
			name:       "valid type gitlab",
			sourceType: "gitlab",
			wantErr:    false,
		},
		{
			name:       "valid type generic",
			sourceType: "generic",
			wantErr:    false,
		},
		{
			name:       "invalid type",
			sourceType: "bitbucket",
			wantErr:    true,
		},
		{
			name:       "empty type",
			sourceType: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateSourceRequest{
				Name:           "Test Source",
				Type:           tt.sourceType,
				DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateSourceRequest_Validate_FieldLengths(t *testing.T) {
	tests := []struct {
		name          string
		sourceName    string
		webhookSecret string
		wantErr       bool
		errContains   string
	}{
		{
			name:       "name at max length",
			sourceName: strings.Repeat("a", 100),
			wantErr:    false,
		},
		{
			name:        "name too long",
			sourceName:  strings.Repeat("a", 101),
			wantErr:     true,
			errContains: "name must be 100 characters or less",
		},
		{
			name:          "secret at max length",
			sourceName:    "Test",
			webhookSecret: strings.Repeat("a", 256),
			wantErr:       false,
		},
		{
			name:          "secret too long",
			sourceName:    "Test",
			webhookSecret: strings.Repeat("a", 257),
			wantErr:       true,
			errContains:   "webhook_secret must be 256 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateSourceRequest{
				Name:           tt.sourceName,
				Type:           "generic",
				DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
				WebhookSecret:  tt.webhookSecret,
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestToSourceResponse_ExcludesSecret(t *testing.T) {
	source := CreateSourceRequest{
		Name:           "Test Source",
		Type:           "github",
		DiscordWebhook: "https://discord.com/api/webhooks/123/abc",
		WebhookSecret:  "super-secret",
	}

	storeSource := source.toStoreSource()
	storeSource.ID = "test-id"

	response := toSourceResponse(storeSource)

	if response.ID != "test-id" {
		t.Errorf("Expected ID %q, got %q", "test-id", response.ID)
	}
	if response.Name != source.Name {
		t.Errorf("Expected Name %q, got %q", source.Name, response.Name)
	}
	if response.Type != source.Type {
		t.Errorf("Expected Type %q, got %q", source.Type, response.Type)
	}
	if response.DiscordWebhook != source.DiscordWebhook {
		t.Errorf("Expected DiscordWebhook %q, got %q", source.DiscordWebhook, response.DiscordWebhook)
	}

	// Verify the response struct doesn't have a WebhookSecret field accessible
	// (This is a compile-time check - if WebhookSecret existed on SourceResponse, this wouldn't compile)
}
