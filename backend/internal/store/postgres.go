package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:    db,
		limit: limit,
	}

	// Run migrations
	if err := store.migrate(); err != nil {
		_ = db.Close()
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
	defer func() { _ = tx.Rollback() }()

	var uid interface{}
	if d.UserID != "" {
		uid = d.UserID
	}

	// Insert new delivery with subsource_id and user_id
	_, err = tx.Exec(`
		INSERT INTO deliveries (event_id, source, title, url, status, subsource_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, d.EventID, d.Source, d.Title, d.URL, d.Status, d.SubsourceID, uid, d.CreatedAt, d.UpdatedAt)

	if err != nil {
		log.Printf("Failed to insert delivery: %v", err)
		return
	}

	// Enforce per-user delivery limit (skip per-user trim if user_id missing — legacy)
	if d.UserID != "" {
		_, err = tx.Exec(`
			DELETE FROM deliveries
			WHERE event_id IN (
				SELECT event_id FROM (
					SELECT event_id FROM deliveries
					WHERE user_id = $1
					ORDER BY created_at DESC
					OFFSET $2
				) AS trim_user_deliveries
			)
		`, d.UserID, s.limit)
	} else {
		_, err = tx.Exec(`
			DELETE FROM deliveries
			WHERE event_id IN (
				SELECT event_id FROM (
					SELECT event_id FROM deliveries
					ORDER BY created_at DESC
					OFFSET $1
				) AS trim_all_deliveries
			)
		`, s.limit)
	}

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

// GetDelivery returns a delivery by event_id (unscoped; for internal/webhook broadcast).
func (s *PostgresStore) GetDelivery(eventID string) (Delivery, bool) {
	var d Delivery
	var subsourceID, uid sql.NullString
	err := s.db.QueryRow(`
		SELECT d.event_id, d.source, d.title, d.url, d.status, d.subsource_id, d.user_id, d.created_at, d.updated_at, d.retry_count, COALESCE(d.last_error, ''),
			COALESCE(ss.name, ''), COALESCE(ss.identifier, '')
		FROM deliveries d
		LEFT JOIN subsources ss ON d.subsource_id = ss.id
		WHERE d.event_id = $1
	`, eventID).Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status, &subsourceID, &uid, &d.CreatedAt, &d.UpdatedAt, &d.RetryCount, &d.LastError, &d.SubsourceName, &d.SubsourceIdentifier)
	if err == sql.ErrNoRows {
		return Delivery{}, false
	}
	if err != nil {
		log.Printf("Failed to get delivery: %v", err)
		return Delivery{}, false
	}
	if subsourceID.Valid {
		d.SubsourceID = &subsourceID.String
	}
	if uid.Valid {
		d.UserID = uid.String
	}
	return d, true
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

// List returns deliveries for a user ordered by created_at descending, limited by the store's limit
func (s *PostgresStore) List(userID string) []Delivery {
	rows, err := s.db.Query(`
		SELECT d.event_id, d.source, d.title, d.url, d.status, d.subsource_id, d.user_id, d.created_at, d.updated_at, d.retry_count, COALESCE(d.last_error, ''),
			COALESCE(s.name, ''), COALESCE(s.identifier, '')
		FROM deliveries d
		LEFT JOIN subsources s ON d.subsource_id = s.id
		WHERE d.user_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2
	`, userID, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries: %v", err)
		return []Delivery{}
	}
	defer func() { _ = rows.Close() }()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceID, uid sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceID, &uid, &d.CreatedAt, &d.UpdatedAt, &d.RetryCount, &d.LastError, &d.SubsourceName, &d.SubsourceIdentifier)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}

		if subsourceID.Valid {
			d.SubsourceID = &subsourceID.String
		}
		if uid.Valid {
			d.UserID = uid.String
		}

		deliveries = append(deliveries, d)
	}

	if deliveries == nil {
		return []Delivery{}
	}

	return deliveries
}

// ListDeliveriesBySubsource returns deliveries filtered by subsource_id and user
func (s *PostgresStore) ListDeliveriesBySubsource(userID string, subsourceID string) []Delivery {
	rows, err := s.db.Query(`
		SELECT d.event_id, d.source, d.title, d.url, d.status, d.subsource_id, d.user_id, d.created_at, d.updated_at, d.retry_count, COALESCE(d.last_error, ''),
			COALESCE(s.name, ''), COALESCE(s.identifier, '')
		FROM deliveries d
		JOIN subsources s ON d.subsource_id = s.id
		JOIN platforms p ON s.platform_id = p.id
		WHERE d.subsource_id = $1 AND p.user_id = $2
		ORDER BY d.created_at DESC
		LIMIT $3
	`, subsourceID, userID, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries by subsource: %v", err)
		return []Delivery{}
	}
	defer func() { _ = rows.Close() }()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceIDVal, uid sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceIDVal, &uid, &d.CreatedAt, &d.UpdatedAt, &d.RetryCount, &d.LastError, &d.SubsourceName, &d.SubsourceIdentifier)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}

		if subsourceIDVal.Valid {
			d.SubsourceID = &subsourceIDVal.String
		}
		if uid.Valid {
			d.UserID = uid.String
		}

		deliveries = append(deliveries, d)
	}

	if deliveries == nil {
		return []Delivery{}
	}

	return deliveries
}

// ListDeliveriesByPlatform returns deliveries filtered by platform_id via subsources JOIN
func (s *PostgresStore) ListDeliveriesByPlatform(userID string, platformID string) []Delivery {
	rows, err := s.db.Query(`
		SELECT d.event_id, d.source, d.title, d.url, d.status, d.subsource_id, d.user_id, d.created_at, d.updated_at, d.retry_count, COALESCE(d.last_error, ''),
			COALESCE(s.name, ''), COALESCE(s.identifier, '')
		FROM deliveries d
		JOIN subsources s ON d.subsource_id = s.id
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.platform_id = $1 AND p.user_id = $2
		ORDER BY d.created_at DESC
		LIMIT $3
	`, platformID, userID, s.limit)

	if err != nil {
		log.Printf("Failed to list deliveries by platform: %v", err)
		return []Delivery{}
	}
	defer func() { _ = rows.Close() }()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		var subsourceID, uid sql.NullString
		err := rows.Scan(&d.EventID, &d.Source, &d.Title, &d.URL, &d.Status,
			&subsourceID, &uid, &d.CreatedAt, &d.UpdatedAt, &d.RetryCount, &d.LastError, &d.SubsourceName, &d.SubsourceIdentifier)
		if err != nil {
			log.Printf("Failed to scan delivery: %v", err)
			continue
		}

		if subsourceID.Valid {
			d.SubsourceID = &subsourceID.String
		}
		if uid.Valid {
			d.UserID = uid.String
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
	defer func() { _ = rows.Close() }()

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
func (s *PostgresStore) AddPlatform(userID string, platform Platform) error {
	normalizePlatformCombines(&platform)
	platform.UserID = userID

	// Generate UUID if not provided
	if platform.ID == "" {
		platform.ID = uuid.New().String()
	}

	// Set creation timestamp if zero-value
	if platform.CreatedAt.IsZero() {
		platform.CreatedAt = time.Now().UTC()
	}

	var webhookSecret sql.NullString
	if platform.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: platform.WebhookSecret, Valid: true}
	}

	var discordWH interface{}
	if strings.TrimSpace(platform.DiscordWebhook) != "" {
		discordWH = platform.DiscordWebhook
	}

	var uid interface{}
	if userID != "" {
		uid = userID
	}

	_, err := s.db.Exec(`
		INSERT INTO platforms (id, user_id, name, discord_webhook, webhook_secret, created_at, filter_include_combine, filter_exclude_combine)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, platform.ID, uid, platform.Name, discordWH, webhookSecret, platform.CreatedAt,
		platform.FilterIncludeCombine, platform.FilterExcludeCombine)

	if err != nil {
		return fmt.Errorf("failed to insert platform: %w", err)
	}

	return nil
}

// ListPlatforms returns platform configurations for a user ordered by name ascending
func (s *PostgresStore) ListPlatforms(userID string) []Platform {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, discord_webhook, webhook_secret, created_at, filter_include_combine, filter_exclude_combine
		FROM platforms
		WHERE user_id = $1
		ORDER BY name ASC
	`, userID)

	if err != nil {
		log.Printf("Failed to list platforms: %v", err)
		return []Platform{}
	}
	defer func() { _ = rows.Close() }()

	var platforms []Platform
	for rows.Next() {
		var p Platform
		var webhookSecret, uid sql.NullString
		var discordWH sql.NullString

		err := rows.Scan(&p.ID, &uid, &p.Name, &discordWH, &webhookSecret, &p.CreatedAt,
			&p.FilterIncludeCombine, &p.FilterExcludeCombine)
		if err != nil {
			log.Printf("Failed to scan platform: %v", err)
			continue
		}

		if uid.Valid {
			p.UserID = uid.String
		}
		p.DiscordWebhook = discordWH.String
		p.WebhookSecret = webhookSecret.String
		normalizePlatformCombines(&p)
		platforms = append(platforms, p)
	}

	// Return empty slice (not nil) if no results
	if platforms == nil {
		return []Platform{}
	}

	return platforms
}

func (s *PostgresStore) scanPlatformRow(row *sql.Row) (Platform, bool) {
	var p Platform
	var webhookSecret, uid sql.NullString
	var discordWH sql.NullString

	err := row.Scan(&p.ID, &uid, &p.Name, &discordWH, &webhookSecret, &p.CreatedAt,
		&p.FilterIncludeCombine, &p.FilterExcludeCombine)
	if err == sql.ErrNoRows {
		return Platform{}, false
	}
	if err != nil {
		log.Printf("Failed to scan platform: %v", err)
		return Platform{}, false
	}
	if uid.Valid {
		p.UserID = uid.String
	}
	p.DiscordWebhook = discordWH.String
	p.WebhookSecret = webhookSecret.String
	normalizePlatformCombines(&p)
	return p, true
}

// GetPlatform retrieves a platform by ID for a user
func (s *PostgresStore) GetPlatform(userID string, id string) (Platform, bool) {
	row := s.db.QueryRow(`
		SELECT id, user_id, name, discord_webhook, webhook_secret, created_at, filter_include_combine, filter_exclude_combine
		FROM platforms
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	p, ok := s.scanPlatformRow(row)
	return p, ok
}

// GetPlatformUnscoped retrieves a platform by ID (internal / worker)
func (s *PostgresStore) GetPlatformUnscoped(id string) (Platform, bool) {
	row := s.db.QueryRow(`
		SELECT id, user_id, name, discord_webhook, webhook_secret, created_at, filter_include_combine, filter_exclude_combine
		FROM platforms
		WHERE id = $1
	`, id)
	p, ok := s.scanPlatformRow(row)
	return p, ok
}

// GetPlatformByName retrieves a platform by name for a user
func (s *PostgresStore) GetPlatformByName(userID string, name string) (Platform, bool) {
	row := s.db.QueryRow(`
		SELECT id, user_id, name, discord_webhook, webhook_secret, created_at, filter_include_combine, filter_exclude_combine
		FROM platforms
		WHERE user_id = $1 AND name = $2
	`, userID, name)
	p, ok := s.scanPlatformRow(row)
	return p, ok
}

// UpdatePlatform modifies platform configuration while preserving created_at
func (s *PostgresStore) UpdatePlatform(userID string, id string, platform Platform) error {
	normalizePlatformCombines(&platform)

	var webhookSecret sql.NullString
	if platform.WebhookSecret != "" {
		webhookSecret = sql.NullString{String: platform.WebhookSecret, Valid: true}
	}

	var discordWH interface{}
	if strings.TrimSpace(platform.DiscordWebhook) != "" {
		discordWH = platform.DiscordWebhook
	}

	result, err := s.db.Exec(`
		UPDATE platforms
		SET discord_webhook = $1, webhook_secret = $2, filter_include_combine = $3, filter_exclude_combine = $4
		WHERE id = $5 AND user_id = $6
	`, discordWH, webhookSecret, platform.FilterIncludeCombine, platform.FilterExcludeCombine, id, userID)

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
func (s *PostgresStore) DeletePlatform(userID string, id string) error {
	result, err := s.db.Exec(`
		DELETE FROM platforms
		WHERE id = $1 AND user_id = $2
	`, id, userID)

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
func (s *PostgresStore) AddSubsource(userID string, subsource Subsource) error {
	if _, ok := s.GetPlatform(userID, subsource.PlatformID); !ok {
		return fmt.Errorf("platform not found: %s", subsource.PlatformID)
	}
	subsource.UserID = userID

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
		platform, found := s.GetPlatformUnscoped(subsource.PlatformID)
		if found {
			subsource.URL = generateSubsourceURL(platform.Name, subsource.Identifier)
		}
	}

	var url sql.NullString
	if subsource.URL != "" {
		url = sql.NullString{String: subsource.URL, Valid: true}
	}

	var uid interface{} = userID
	if userID == "" {
		uid = nil
	}

	_, err := s.db.Exec(`
		INSERT INTO subsources (id, user_id, platform_id, name, identifier, url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, subsource.ID, uid, subsource.PlatformID, subsource.Name, subsource.Identifier, url, subsource.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert subsource: %w", err)
	}

	return nil
}

// ListSubsources returns subsources for a platform with JOIN to platforms, ordered by created_at DESC
func (s *PostgresStore) ListSubsources(userID string, platformID string) []SubsourceWithPlatform {
	rows, err := s.db.Query(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name, p.user_id::text
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.platform_id = $1 AND p.user_id = $2
		ORDER BY s.created_at DESC
	`, platformID, userID)

	if err != nil {
		log.Printf("Failed to list subsources: %v", err)
		return []SubsourceWithPlatform{}
	}
	defer func() { _ = rows.Close() }()

	var subsources []SubsourceWithPlatform
	for rows.Next() {
		var s SubsourceWithPlatform
		var url sql.NullString

		var ownerID sql.NullString
		err := rows.Scan(&s.ID, &s.PlatformID, &s.Name, &s.Identifier, &url, &s.CreatedAt, &s.PlatformName, &ownerID)
		if err != nil {
			log.Printf("Failed to scan subsource: %v", err)
			continue
		}

		// Convert NULL field to empty string
		s.URL = url.String
		if ownerID.Valid {
			s.UserID = ownerID.String
		}
		subsources = append(subsources, s)
	}

	// Return empty slice (not nil) if no results
	if subsources == nil {
		return []SubsourceWithPlatform{}
	}

	return subsources
}

// ListAllSubsources returns user-owned subsources (platform.user_id NOT NULL), ordered by created_at DESC
func (s *PostgresStore) ListAllSubsources() []SubsourceWithPlatform {
	rows, err := s.db.Query(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name, p.user_id::text
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE p.user_id IS NOT NULL
		ORDER BY s.created_at DESC
	`)

	if err != nil {
		log.Printf("Failed to list all subsources: %v", err)
		return []SubsourceWithPlatform{}
	}
	defer func() { _ = rows.Close() }()

	var subsources []SubsourceWithPlatform
	for rows.Next() {
		var s SubsourceWithPlatform
		var url sql.NullString

		var ownerID sql.NullString
		err := rows.Scan(&s.ID, &s.PlatformID, &s.Name, &s.Identifier, &url, &s.CreatedAt, &s.PlatformName, &ownerID)
		if err != nil {
			log.Printf("Failed to scan subsource: %v", err)
			continue
		}

		// Convert NULL field to empty string
		s.URL = url.String
		if ownerID.Valid {
			s.UserID = ownerID.String
		}
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
	var ownerID sql.NullString

	err := s.db.QueryRow(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name, p.user_id::text
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.id = $1
	`, id).Scan(&sub.ID, &sub.PlatformID, &sub.Name, &sub.Identifier, &url, &sub.CreatedAt, &sub.PlatformName, &ownerID)

	if err == sql.ErrNoRows {
		return SubsourceWithPlatform{}, false
	}

	if err != nil {
		log.Printf("Failed to get subsource: %v", err)
		return SubsourceWithPlatform{}, false
	}

	sub.URL = url.String
	if ownerID.Valid {
		sub.UserID = ownerID.String
	}

	return sub, true
}

// GetSubsourceForUser returns a subsource only if its platform belongs to the user
func (s *PostgresStore) GetSubsourceForUser(userID string, id string) (SubsourceWithPlatform, bool) {
	var sub SubsourceWithPlatform
	var url sql.NullString
	var ownerID sql.NullString

	err := s.db.QueryRow(`
		SELECT s.id, s.platform_id, s.name, s.identifier, s.url, s.created_at, p.name AS platform_name, p.user_id::text
		FROM subsources s
		JOIN platforms p ON s.platform_id = p.id
		WHERE s.id = $1 AND p.user_id = $2
	`, id, userID).Scan(&sub.ID, &sub.PlatformID, &sub.Name, &sub.Identifier, &url, &sub.CreatedAt, &sub.PlatformName, &ownerID)

	if err == sql.ErrNoRows {
		return SubsourceWithPlatform{}, false
	}

	if err != nil {
		log.Printf("Failed to get subsource: %v", err)
		return SubsourceWithPlatform{}, false
	}

	sub.URL = url.String
	if ownerID.Valid {
		sub.UserID = ownerID.String
	}

	return sub, true
}

// UpdateSubsource modifies subsource configuration while preventing platform_id changes
func (s *PostgresStore) UpdateSubsource(userID string, id string, subsource Subsource) error {
	var url sql.NullString
	if subsource.URL != "" {
		url = sql.NullString{String: subsource.URL, Valid: true}
	}

	result, err := s.db.Exec(`
		UPDATE subsources s
		SET name = $1, identifier = $2, url = $3
		FROM platforms p
		WHERE s.id = $4 AND s.platform_id = p.id AND p.user_id = $5
	`, subsource.Name, subsource.Identifier, url, id, userID)

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
func (s *PostgresStore) DeleteSubsource(userID string, id string) error {
	result, err := s.db.Exec(`
		DELETE FROM subsources s
		USING platforms p
		WHERE s.id = $1 AND s.platform_id = p.id AND p.user_id = $2
	`, id, userID)

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

// AddFilter inserts a new destination filter for a platform
func (s *PostgresStore) AddFilter(userID string, filter DestinationFilter) error {
	if _, ok := s.GetPlatform(userID, filter.PlatformID); !ok {
		return fmt.Errorf("platform not found: %s", filter.PlatformID)
	}
	filter.UserID = userID

	if filter.ID == "" {
		filter.ID = uuid.New().String()
	}
	if filter.CreatedAt.IsZero() {
		filter.CreatedAt = time.Now().UTC()
	}

	var uid interface{} = userID
	if userID == "" {
		uid = nil
	}

	_, err := s.db.Exec(`
		INSERT INTO destination_filters (id, user_id, platform_id, filter_type, pattern, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, filter.ID, uid, filter.PlatformID, filter.FilterType, filter.Pattern, filter.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert filter: %w", err)
	}
	return nil
}

// ListFilters returns all destination filters for a given platform
func (s *PostgresStore) ListFilters(userID string, platformID string) []DestinationFilter {
	rows, err := s.db.Query(`
		SELECT f.id, f.platform_id, f.filter_type, f.pattern, f.created_at
		FROM destination_filters f
		JOIN platforms p ON f.platform_id = p.id
		WHERE f.platform_id = $1 AND p.user_id = $2
		ORDER BY f.created_at DESC
	`, platformID, userID)
	if err != nil {
		log.Printf("Failed to list filters: %v", err)
		return []DestinationFilter{}
	}
	defer func() { _ = rows.Close() }()

	var filters []DestinationFilter
	for rows.Next() {
		var f DestinationFilter
		if err := rows.Scan(&f.ID, &f.PlatformID, &f.FilterType, &f.Pattern, &f.CreatedAt); err != nil {
			log.Printf("Failed to scan filter: %v", err)
			continue
		}
		filters = append(filters, f)
	}
	if filters == nil {
		return []DestinationFilter{}
	}
	return filters
}

// DeleteFilter removes a destination filter by ID
func (s *PostgresStore) DeleteFilter(userID string, id string) error {
	result, err := s.db.Exec(`
		DELETE FROM destination_filters f
		USING platforms p
		WHERE f.id = $1 AND f.platform_id = p.id AND p.user_id = $2
	`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete filter: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("filter not found: %s", id)
	}
	return nil
}

// CreateUser inserts a new user into the database
func (s *PostgresStore) CreateUser(user User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)`,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by email address
func (s *PostgresStore) GetUserByEmail(email string) (User, bool) {
	var u User
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, created_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

// GetUserByID retrieves a user by ID
func (s *PostgresStore) GetUserByID(id string) (User, bool) {
	var u User
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, created_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

// CreateSession inserts a new session into the database
func (s *PostgresStore) CreateSession(session Session) error {
	if session.SessionID == "" {
		session.SessionID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO sessions (session_id, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)`,
		session.SessionID, session.UserID, session.CreatedAt, session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by session ID
func (s *PostgresStore) GetSession(sessionID string) (Session, bool) {
	var sess Session
	err := s.db.QueryRow(`
		SELECT session_id, user_id, created_at, expires_at
		FROM sessions WHERE session_id = $1`, sessionID,
	).Scan(&sess.SessionID, &sess.UserID, &sess.CreatedAt, &sess.ExpiresAt)
	if err != nil {
		return Session{}, false
	}
	return sess, true
}

// DeleteSession removes a session by session ID
func (s *PostgresStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE session_id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes all sessions past their expiry time
func (s *PostgresStore) DeleteExpiredSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}