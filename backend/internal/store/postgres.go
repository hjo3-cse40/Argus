package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresStore implements the Store interface using PostgreSQL
type PostgresStore struct {
	db    *sql.DB
	limit int
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(connStr string, limit int) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connectivity
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:    db,
		limit: limit,
	}

	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	log.Println("PostgreSQL store initialized successfully")
	return store, nil
}

// Close releases all database connections
func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate executes idempotent database migrations
func (s *PostgresStore) migrate() error {
	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// AddQueued inserts a new delivery record with status 'queued' and enforces the delivery limit
func (s *PostgresStore) AddQueued(d Delivery) {
	now := time.Now().UTC()
	d.Status = StatusQueued
	d.CreatedAt = now
	d.UpdatedAt = now

	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// Insert new delivery with subsource_id
	_, err = tx.Exec(`
		INSERT INTO deliveries (event_id, source, title, url, status, subsource_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, d.EventID, d.Source, d.Title, d.URL, d.Status, d.SubsourceID, d.CreatedAt, d.UpdatedAt)

	if err != nil {
		log.Printf("Failed to insert delivery: %v", err)
		return
	}

	// Enforce limit by deleting oldest records
	_, err = tx.Exec(`
		DELETE FROM deliveries
		WHERE event_id IN (
			SELECT event_id FROM deliveries
			ORDER BY created_at DESC
			OFFSET $1
		)
	`, s.limit)

	if err != nil {
		log.Printf("Failed to enforce delivery limit: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}
}

// MarkDelivered updates a delivery's status to 'delivered' and sets updated_at to current UTC time
func (s *PostgresStore) MarkDelivered(eventID string) bool {
	result, err := s.db.Exec(`
		UPDATE deliveries
		SET status = $1, updated_at = $2
		WHERE event_id = $3
	`, StatusDelivered, time.Now().UTC(), eventID)

	if err != nil {
		log.Printf("Failed to mark delivery as delivered: %v", err)
		return false
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		return false
	}

	return rows > 0
}

// MarkFailed updates a delivery's status to 'failed' and sets retry_count and last_error
func (s *PostgresStore) MarkFailed(eventID string, retryCount int, lastErr string) bool {
	result, err := s.db.Exec(`
		UPDATE deliveries
		SET status = $1, updated_at = $2, retry_count = $4, last_error = $5
		WHERE event_id = $3
	`, StatusFailed, time.Now().UTC(), eventID, retryCount, lastErr)
	if err != nil {
		log.Printf("Failed to mark delivery as failed: %v", err)
		return false
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false
	}
	return rows > 0
}

// List returns all deliveries ordered by created_at descending, limited by the store's limit
func (s *PostgresStore) List() []Delivery {
	rows, err := s.db.Query(`
		SELECT event_id, source, title, url, status, subsource_id, created_at, updated_at
		FROM deliveries
		ORDER BY created_at DESC
		LIMIT $1
	`, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries: %v", err)
		return []Delivery{}
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceID sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}
		
		// Convert NULL subsource_id to nil pointer
		if subsourceID.Valid {
			d.SubsourceID = &subsourceID.String
		}
		
		deliveries = append(deliveries, d)
	}

	if deliveries == nil {
		return []Delivery{}
	}

	return deliveries
}

// ListDeliveriesBySubsource returns deliveries filtered by subsource_id
func (s *PostgresStore) ListDeliveriesBySubsource(subsourceID string) []Delivery {
	rows, err := s.db.Query(`
		SELECT event_id, source, title, url, status, subsource_id, created_at, updated_at
		FROM deliveries
		WHERE subsource_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, subsourceID, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries by subsource: %v", err)
		return []Delivery{}
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceIDVal sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceIDVal, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}
		
		// Convert NULL subsource_id to nil pointer
		if subsourceIDVal.Valid {
			d.SubsourceID = &subsourceIDVal.String
		}
		
		deliveries = append(deliveries, d)
	}

	if deliveries == nil {
		return []Delivery{}
	}

	return deliveries
}

// ListDeliveriesByPlatform returns deliveries filtered by platform_id via subsources JOIN
func (s *PostgresStore) ListDeliveriesByPlatform(platformID string) []Delivery {
	rows, err := s.db.Query(`
		SELECT d.event_id, d.source, d.title, d.url, d.status, d.subsource_id, d.created_at, d.updated_at
		FROM deliveries d
		JOIN subsources s ON d.subsource_id = s.id
		WHERE s.platform_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2
	`, platformID, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries by platform: %v", err)
		return []Delivery{}
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceID sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}
		
		// Convert NULL subsource_id to nil pointer
		if subsourceID.Valid {
			d.SubsourceID = &subsourceID.String
		}
		
		deliveries = append(deliveries, d)
	}

	if deliveries == nil {
		return []Delivery{}
	}

	return deliveries
}

// AddSource inserts a new source configuration with generated UUID if ID is empty
func (s *PostgresStore) AddSource(source Source) error {
	// Generate UUID if not provided
	if source.ID == "" {
		source.ID = uuid.New().String()
	}

	// Set creation timestamp if zero-value
	if source.CreatedAt.IsZero() {
		source.CreatedAt = time.Now().UTC()
	}

	// Use sql.NullString for optional fields
	var repoURL, webhookSecret sql.NullString
	if source.RepositoryURL != "" {
		repoURL = sql.NullString{String: source.RepositoryURL, Valid: true}
	}
	if source.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: source.WebhookSecret, Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO sources (id, name, type, repository_url, discord_webhook, webhook_secret, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, source.ID, source.Name, source.Type, repoURL, source.DiscordWebhook, webhookSecret, source.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert source: %w", err)
	}

	return nil
}

// ListSources returns all source configurations ordered by created_at descending
func (s *PostgresStore) ListSources() []Source {
	rows, err := s.db.Query(`
		SELECT id, name, type, repository_url, discord_webhook, webhook_secret, created_at
		FROM sources
		ORDER BY created_at DESC
	`)

	if err != nil {
		log.Printf("Failed to list sources: %v", err)
		return []Source{}
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var src Source
		var repoURL, webhookSecret sql.NullString

		err := rows.Scan(&src.ID, &src.Name, &src.Type, &repoURL,
			&src.DiscordWebhook, &webhookSecret, &src.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan source: %v", err)
			continue
		}

		// Convert NULL fields to empty strings
		src.RepositoryURL = repoURL.String
		src.WebhookSecret = webhookSecret.String
		sources = append(sources, src)
	}

	// Return empty slice (not nil) if no results
	if sources == nil {
		return []Source{}
	}

	return sources
}

// GetSource retrieves a source by ID
func (s *PostgresStore) GetSource(id string) (Source, bool) {
	var src Source
	var repoURL, webhookSecret sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, type, repository_url, discord_webhook, webhook_secret, created_at
		FROM sources
		WHERE id = $1
	`, id).Scan(&src.ID, &src.Name, &src.Type, &repoURL,
		&src.DiscordWebhook, &webhookSecret, &src.CreatedAt)

	if err == sql.ErrNoRows {
		return Source{}, false
	}

	if err != nil {
		log.Printf("Failed to get source: %v", err)
		return Source{}, false
	}

	// Convert NULL fields to empty strings
	src.RepositoryURL = repoURL.String
	src.WebhookSecret = webhookSecret.String

	return src, true
}

// GetSourceByName retrieves a source by name using indexed column
func (s *PostgresStore) GetSourceByName(name string) (Source, bool) {
	var src Source
	var repoURL, webhookSecret sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, type, repository_url, discord_webhook, webhook_secret, created_at
		FROM sources
		WHERE name = $1
	`, name).Scan(&src.ID, &src.Name, &src.Type, &repoURL,
		&src.DiscordWebhook, &webhookSecret, &src.CreatedAt)

	if err == sql.ErrNoRows {
		return Source{}, false
	}

	if err != nil {
		log.Printf("Failed to get source by name: %v", err)
		return Source{}, false
	}

	// Convert NULL fields to empty strings
	src.RepositoryURL = repoURL.String
	src.WebhookSecret = webhookSecret.String

	return src, true
}

// AddPlatform inserts a new platform configuration with generated UUID if ID is empty
func (s *PostgresStore) AddPlatform(platform Platform) error {
	// Generate UUID if not provided
	if platform.ID == "" {
		platform.ID = uuid.New().String()
	}

	// Set creation timestamp if zero-value
	if platform.CreatedAt.IsZero() {
		platform.CreatedAt = time.Now().UTC()
	}

	// Use sql.NullString for optional webhook_secret field
	var webhookSecret sql.NullString
	if platform.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: platform.WebhookSecret, Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO platforms (id, name, discord_webhook, webhook_secret, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, platform.ID, platform.Name, platform.DiscordWebhook, webhookSecret, platform.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert platform: %w", err)
	}

	return nil
}

// ListPlatforms returns all platform configurations ordered by name ascending
func (s *PostgresStore) ListPlatforms() []Platform {
	rows, err := s.db.Query(`
		SELECT id, name, discord_webhook, webhook_secret, created_at
		FROM platforms
		ORDER BY name ASC
	`)

	if err != nil {
		log.Printf("Failed to list platforms: %v", err)
		return []Platform{}
	}
	defer rows.Close()

	var platforms []Platform
	for rows.Next() {
		var p Platform
		var webhookSecret sql.NullString

		err := rows.Scan(&p.ID, &p.Name, &p.DiscordWebhook, &webhookSecret, &p.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan platform: %v", err)
			continue
		}

		// Convert NULL field to empty string
		p.WebhookSecret = webhookSecret.String
		platforms = append(platforms, p)
	}

	// Return empty slice (not nil) if no results
	if platforms == nil {
		return []Platform{}
	}

	return platforms
}

// GetPlatform retrieves a platform by ID
func (s *PostgresStore) GetPlatform(id string) (Platform, bool) {
	var p Platform
	var webhookSecret sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, discord_webhook, webhook_secret, created_at
		FROM platforms
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.DiscordWebhook, &webhookSecret, &p.CreatedAt)

	if err == sql.ErrNoRows {
		return Platform{}, false
	}

	if err != nil {
		log.Printf("Failed to get platform: %v", err)
		return Platform{}, false
	}

	// Convert NULL field to empty string
	p.WebhookSecret = webhookSecret.String

	return p, true
}

// GetPlatformByName retrieves a platform by name using indexed column
func (s *PostgresStore) GetPlatformByName(name string) (Platform, bool) {
	var p Platform
	var webhookSecret sql.NullString

	err := s.db.QueryRow(`
		SELECT id, name, discord_webhook, webhook_secret, created_at
		FROM platforms
		WHERE name = $1
	`, name).Scan(&p.ID, &p.Name, &p.DiscordWebhook, &webhookSecret, &p.CreatedAt)

	if err == sql.ErrNoRows {
		return Platform{}, false
	}

	if err != nil {
		log.Printf("Failed to get platform by name: %v", err)
		return Platform{}, false
	}

	// Convert NULL field to empty string
	p.WebhookSecret = webhookSecret.String

	return p, true
}

// UpdatePlatform modifies platform configuration while preserving created_at
func (s *PostgresStore) UpdatePlatform(id string, platform Platform) error {
	// Use sql.NullString for optional webhook_secret field
	var webhookSecret sql.NullString
	if platform.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: platform.WebhookSecret, Valid: true}
	}

	result, err := s.db.Exec(`
		UPDATE platforms
		SET discord_webhook = $1, webhook_secret = $2
		WHERE id = $3
	`, platform.DiscordWebhook, webhookSecret, id)

	if err != nil {
		return fmt.Errorf("failed to update platform: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("platform not found: %s", id)
	}

	return nil
}

// DeletePlatform removes a platform and cascades to subsources
func (s *PostgresStore) DeletePlatform(id string) error {
	result, err := s.db.Exec(`
		DELETE FROM platforms
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete platform: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("platform not found: %s", id)
	}

	return nil
}

// generateSubsourceURL creates a URL based on platform name and identifier
func generateSubsourceURL(platformName, identifier string) string {
	switch platformName {
	case "youtube":
		return fmt.Sprintf("https://youtube.com/channel/%s", identifier)
	case "reddit":
		return fmt.Sprintf("https://reddit.com/r/%s", identifier)
	case "x":
		return fmt.Sprintf("https://x.com/%s", identifier)
	default:
		return ""
	}
}

// AddSubsource inserts a new subsource configuration with generated UUID, timestamp, and URL auto-generation
func (s *PostgresStore) AddSubsource(subsource Subsource) error {
	// Generate UUID if not provided
	if subsource.ID == "" {
		subsource.ID = uuid.New().String()
	}

	// Set creation timestamp if zero-value
	if subsource.CreatedAt.IsZero() {
		subsource.CreatedAt = time.Now().UTC()
	}

	// Auto-generate URL if not provided
	if subsource.URL == "" {
		// Get platform to determine URL format
		platform, found := s.GetPlatform(subsource.PlatformID)
		if found {
			subsource.URL = generateSubsourceURL(platform.Name, subsource.Identifier)
		}
	}

	// Use sql.NullString for optional URL field
	var url sql.NullString
	if subsource.URL != "" {
		url = sql.NullString{String: subsource.URL, Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO subsources (id, platform_id, name, identifier, url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, subsource.ID, subsource.PlatformID, subsource.Name, subsource.Identifier, url, subsource.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert subsource: %w", err)
	}

	return nil
}

// ListSubsources returns subsources for a platform with JOIN to platforms, ordered by created_at DESC
func (s *PostgresStore) ListSubsources(platformID string) []SubsourceWithPlatform {
	rows, err := s.db.Query(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.platform_id = $1
		ORDER BY s.created_at DESC
	`, platformID)

	if err != nil {
		log.Printf("Failed to list subsources: %v", err)
		return []SubsourceWithPlatform{}
	}
	defer rows.Close()

	var subsources []SubsourceWithPlatform
	for rows.Next() {
		var s SubsourceWithPlatform
		var url sql.NullString

		err := rows.Scan(&s.ID, &s.PlatformID, &s.Name, &s.Identifier, &url, &s.CreatedAt, &s.PlatformName)
		if err != nil {
			log.Printf("Failed to scan subsource: %v", err)
			continue
		}

		// Convert NULL field to empty string
		s.URL = url.String
		subsources = append(subsources, s)
	}

	// Return empty slice (not nil) if no results
	if subsources == nil {
		return []SubsourceWithPlatform{}
	}

	return subsources
}

// ListAllSubsources returns all subsources with JOIN to platforms, ordered by created_at DESC
func (s *PostgresStore) ListAllSubsources() []SubsourceWithPlatform {
	rows, err := s.db.Query(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		ORDER BY s.created_at DESC
	`)

	if err != nil {
		log.Printf("Failed to list all subsources: %v", err)
		return []SubsourceWithPlatform{}
	}
	defer rows.Close()

	var subsources []SubsourceWithPlatform
	for rows.Next() {
		var s SubsourceWithPlatform
		var url sql.NullString

		err := rows.Scan(&s.ID, &s.PlatformID, &s.Name, &s.Identifier, &url, &s.CreatedAt, &s.PlatformName)
		if err != nil {
			log.Printf("Failed to scan subsource: %v", err)
			continue
		}

		// Convert NULL field to empty string
		s.URL = url.String
		subsources = append(subsources, s)
	}

	// Return empty slice (not nil) if no results
	if subsources == nil {
		return []SubsourceWithPlatform{}
	}

	return subsources
}

// GetSubsource retrieves a subsource with platform information in single query
func (s *PostgresStore) GetSubsource(id string) (SubsourceWithPlatform, bool) {
	var sub SubsourceWithPlatform
	var url sql.NullString

	err := s.db.QueryRow(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.id = $1
	`, id).Scan(&sub.ID, &sub.PlatformID, &sub.Name, &sub.Identifier, &url, &sub.CreatedAt, &sub.PlatformName)

	if err == sql.ErrNoRows {
		return SubsourceWithPlatform{}, false
	}

	if err != nil {
		log.Printf("Failed to get subsource: %v", err)
		return SubsourceWithPlatform{}, false
	}

	// Convert NULL field to empty string
	sub.URL = url.String

	return sub, true
}

// UpdateSubsource modifies subsource configuration while preventing platform_id changes
func (s *PostgresStore) UpdateSubsource(id string, subsource Subsource) error {
	// Use sql.NullString for optional URL field
	var url sql.NullString
	if subsource.URL != "" {
		url = sql.NullString{String: subsource.URL, Valid: true}
	}

	// Update only name, identifier, and url - platform_id is immutable
	result, err := s.db.Exec(`
		UPDATE subsources
		SET name = $1, identifier = $2, url = $3
		WHERE id = $4
	`, subsource.Name, subsource.Identifier, url, id)

	if err != nil {
		return fmt.Errorf("failed to update subsource: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subsource not found: %s", id)
	}

	return nil
}

// DeleteSubsource removes a subsource
func (s *PostgresStore) DeleteSubsource(id string) error {
	result, err := s.db.Exec(`
		DELETE FROM subsources
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete subsource: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subsource not found: %s", id)
	}

	return nil
}
