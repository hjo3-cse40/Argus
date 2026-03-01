package store

import (
	"fmt"
	"time"
)

// migrations contains idempotent SQL statements for database schema initialization
var migrations = []string{
	// Create deliveries table
	`CREATE TABLE IF NOT EXISTS deliveries (
		event_id TEXT PRIMARY KEY,
		source TEXT NOT NULL,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		status TEXT NOT NULL CHECK (status IN ('queued', 'delivered')),
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	)`,

	// Create index on deliveries.status for efficient status-based queries
	`CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries(status)`,

	// Create index on deliveries.created_at for efficient time-ordered retrieval
	`CREATE INDEX IF NOT EXISTS idx_deliveries_created_at ON deliveries(created_at DESC)`,

	// Create sources table
	`CREATE TABLE IF NOT EXISTS sources (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		repository_url TEXT,
		discord_webhook TEXT NOT NULL,
		webhook_secret TEXT,
		created_at TIMESTAMPTZ NOT NULL
	)`,

	// Create unique index on sources.name to prevent duplicate source names
	`CREATE INDEX IF NOT EXISTS idx_sources_name ON sources(name)`,

	// Create platforms table
	`CREATE TABLE IF NOT EXISTS platforms (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL UNIQUE CHECK (name IN ('youtube', 'reddit', 'x')),
		discord_webhook TEXT NOT NULL,
		webhook_secret TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	// Create unique index on platforms.name to prevent duplicate platform names
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_platforms_name ON platforms(name)`,

	// Create subsources table
	`CREATE TABLE IF NOT EXISTS subsources (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		platform_id UUID NOT NULL REFERENCES platforms(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		identifier TEXT NOT NULL,
		url TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(platform_id, identifier)
	)`,

	// Create index on subsources.platform_id for efficient platform-based queries
	`CREATE INDEX IF NOT EXISTS idx_subsources_platform_id ON subsources(platform_id)`,

	// Create unique index on subsources(platform_id, identifier)
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_subsources_platform_identifier ON subsources(platform_id, identifier)`,

	// Add subsource_id column to deliveries table
	`ALTER TABLE deliveries ADD COLUMN IF NOT EXISTS subsource_id UUID REFERENCES subsources(id) ON DELETE SET NULL`,

	// Create index on deliveries.subsource_id for efficient subsource-based queries
	`CREATE INDEX IF NOT EXISTS idx_deliveries_subsource_id ON deliveries(subsource_id)`,
}

// MigrateFlatToHierarchical migrates data from the flat sources table to hierarchical platforms and subsources tables
func MigrateFlatToHierarchical(store Store) error {
	// This migration function transforms the flat sources table into hierarchical structure
	// It should only be called once after the schema is set up but before dropping the sources table
	
	// Step 1: Get all sources from the flat table
	sources := store.ListSources()
	if len(sources) == 0 {
		// No sources to migrate, this is fine
		return nil
	}
	
	// Step 2: Extract unique platform types and validate webhook consistency
	platformData := make(map[string]struct {
		webhook string
		secret  string
		createdAt time.Time
	})
	
	for _, source := range sources {
		if existing, exists := platformData[source.Type]; exists {
			// Check webhook consistency
			if existing.webhook != source.DiscordWebhook {
				return fmt.Errorf("migration failed: inconsistent webhooks detected for platform '%s'", source.Type)
			}
		} else {
			platformData[source.Type] = struct {
				webhook string
				secret  string
				createdAt time.Time
			}{
				webhook:   source.DiscordWebhook,
				secret:    source.WebhookSecret,
				createdAt: source.CreatedAt,
			}
		}
	}
	
	// Step 3: Create platform records for each unique type
	platformIDs := make(map[string]string) // type -> platform_id
	for platformType, data := range platformData {
		// Check if platform already exists (idempotency)
		existingPlatform, found := store.GetPlatformByName(platformType)
		if found {
			platformIDs[platformType] = existingPlatform.ID
			continue
		}
		
		platform := Platform{
			Name:           platformType,
			DiscordWebhook: data.webhook,
			WebhookSecret:  data.secret,
			CreatedAt:      data.createdAt,
		}
		
		if err := store.AddPlatform(platform); err != nil {
			return fmt.Errorf("failed to create platform %s: %w", platformType, err)
		}
		
		// Get the created platform to retrieve its ID
		createdPlatform, found := store.GetPlatformByName(platformType)
		if !found {
			return fmt.Errorf("failed to retrieve created platform: %s", platformType)
		}
		platformIDs[platformType] = createdPlatform.ID
	}
	
	// Step 4: Create subsource records from existing sources
	for _, source := range sources {
		platformID := platformIDs[source.Type]
		
		// Use repository_url as identifier if available, otherwise use name
		identifier := source.Name
		if source.RepositoryURL != "" {
			identifier = source.RepositoryURL
		}
		
		// Check if subsource already exists (idempotency)
		existingSubsources := store.ListSubsources(platformID)
		subsourceExists := false
		for _, existing := range existingSubsources {
			if existing.Identifier == identifier {
				subsourceExists = true
				break
			}
		}
		
		if subsourceExists {
			continue
		}
		
		// Don't set URL field - let AddSubsource auto-generate it
		// The RepositoryURL is just an identifier, not a full URL
		subsource := Subsource{
			PlatformID: platformID,
			Name:       source.Name,
			Identifier: identifier,
			URL:        "", // Leave empty to trigger auto-generation
			CreatedAt:  source.CreatedAt,
		}
		
		if err := store.AddSubsource(subsource); err != nil {
			return fmt.Errorf("failed to create subsource %s: %w", source.Name, err)
		}
	}
	
	// Step 5: Migration completed successfully
	// Note: The actual dropping of the sources table should be done by the database migration
	// This function only handles the data migration part
	
	return nil
}
