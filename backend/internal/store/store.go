package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	StatusQueued    DeliveryStatus = "queued"
	StatusDelivered DeliveryStatus = "delivered"
	StatusFailed    DeliveryStatus = "failed"
)

type Delivery struct {
	EventID     string         `json:"event_id"`
	Source      string         `json:"source"`
	Title       string         `json:"title"`
	URL         string         `json:"url"`
	Status      DeliveryStatus `json:"status"`
	SubsourceID *string        `json:"subsource_id,omitempty"`
	UserID      string         `json:"user_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	RetryCount  int            `json:"retry_count"`
	LastError   string         `json:"last_error,omitempty"`
}

type Source struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	RepositoryURL  string    `json:"repository_url,omitempty"`
	DiscordWebhook string    `json:"discord_webhook"`
	WebhookSecret  string    `json:"webhook_secret,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Platform struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id,omitempty"`
	Name                 string    `json:"name"`
	DiscordWebhook       string    `json:"discord_webhook"`
	WebhookSecret        string    `json:"webhook_secret,omitempty"`
	FilterIncludeCombine string    `json:"filter_include_combine,omitempty"` // "any" (default) or "all"
	FilterExcludeCombine string    `json:"filter_exclude_combine,omitempty"` // "any" (default) or "all"
	CreatedAt            time.Time `json:"created_at"`
}

type Subsource struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id,omitempty"`
	PlatformID string    `json:"platform_id"`
	Name       string    `json:"name"`
	Identifier string    `json:"identifier"`
	URL        string    `json:"url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubsourceWithPlatform struct {
	Subsource
	PlatformName string `json:"platform_name"`
}
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DestinationFilter struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id,omitempty"`
	PlatformID string    `json:"platform_id"`
	FilterType string    `json:"filter_type"`
	Pattern    string    `json:"pattern"`
	CreatedAt  time.Time `json:"created_at"`
}

type MemoryStore struct {
	mu         sync.Mutex
	deliveries []Delivery
	sources    []Source
	platforms  []Platform
	subsources []Subsource
	users      []User
	sessions   []Session
	filters    []DestinationFilter
	limit      int
}

func NewMemoryStore(limit int) *MemoryStore {
	return &MemoryStore{limit: limit}
}

func (s *MemoryStore) AddQueued(d Delivery) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	d.Status = StatusQueued
	d.CreatedAt = now
	d.UpdatedAt = now

	s.deliveries = append([]Delivery{d}, s.deliveries...)

	if d.UserID != "" {
		userCount := 0
		for _, x := range s.deliveries {
			if x.UserID == d.UserID {
				userCount++
			}
		}
		for userCount > s.limit {
			removed := false
			for i := len(s.deliveries) - 1; i >= 0; i-- {
				if s.deliveries[i].UserID == d.UserID {
					s.deliveries = append(s.deliveries[:i], s.deliveries[i+1:]...)
					userCount--
					removed = true
					break
				}
			}
			if !removed {
				break
			}
		}
	} else if len(s.deliveries) > s.limit {
		s.deliveries = s.deliveries[:s.limit]
	}
}

func (s *MemoryStore) MarkDelivered(eventID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.deliveries {
		if s.deliveries[i].EventID == eventID {
			s.deliveries[i].Status = StatusDelivered
			s.deliveries[i].UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

func (s *MemoryStore) List(userID string) []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []Delivery
	for _, d := range s.deliveries {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	if len(out) > s.limit {
		out = out[:s.limit]
	}
	return out
}

func (s *MemoryStore) ListDeliveriesBySubsource(userID string, subsourceID string) []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []Delivery
	for _, d := range s.deliveries {
		if d.UserID != userID {
			continue
		}
		if d.SubsourceID != nil && *d.SubsourceID == subsourceID {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func (s *MemoryStore) ListDeliveriesByPlatform(userID string, platformID string) []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	subsourceIDs := make(map[string]bool)
	for _, sub := range s.subsources {
		if sub.PlatformID == platformID {
			subsourceIDs[sub.ID] = true
		}
	}

	var filtered []Delivery
	for _, d := range s.deliveries {
		if d.UserID != userID {
			continue
		}
		if d.SubsourceID != nil && subsourceIDs[*d.SubsourceID] {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// AddSource stores a new source configuration with generated UUID
func (s *MemoryStore) AddSource(source Source) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate UUID if not provided
	if source.ID == "" {
		source.ID = uuid.New().String()
	}

	// Set creation timestamp
	if source.CreatedAt.IsZero() {
		source.CreatedAt = time.Now().UTC()
	}

	s.sources = append(s.sources, source)
	return nil
}

// ListSources returns all source configurations
func (s *MemoryStore) ListSources() []Source {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Source, len(s.sources))
	copy(out, s.sources)
	return out
}

// GetSource retrieves a source by ID
func (s *MemoryStore) GetSource(id string) (Source, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, source := range s.sources {
		if source.ID == id {
			return source, true
		}
	}
	return Source{}, false
}

// GetSourceByName retrieves a source by name (for worker routing)
func (s *MemoryStore) GetSourceByName(name string) (Source, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, source := range s.sources {
		if source.Name == name {
			return source, true
		}
	}
	return Source{}, false
}

// MarkFailed records failure
func (s *MemoryStore) GetDelivery(eventID string) (Delivery, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, d := range s.deliveries {
		if d.EventID == eventID {
			return d, true
		}
	}
	return Delivery{}, false
}

func (s *MemoryStore) MarkFailed(eventID string, retryCount int, lastErr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.deliveries {
		if s.deliveries[i].EventID == eventID {
			s.deliveries[i].Status = StatusFailed
			s.deliveries[i].RetryCount = retryCount
			s.deliveries[i].LastError = lastErr
			s.deliveries[i].UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

// Close is a no-op for MemoryStore (implements Store interface)
func (s *MemoryStore) Close() error {
	return nil
}

// AddPlatform stores a new platform configuration with generated UUID
func (s *MemoryStore) AddPlatform(userID string, platform Platform) error {
	normalizePlatformCombines(&platform)
	platform.UserID = userID
	// Validate platform data
	if err := validatePlatform(platform); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.platforms {
		if p.UserID == userID && p.Name == platform.Name {
			return &ValidationError{Details: []string{fmt.Sprintf("platform name already exists: %s", platform.Name)}}
		}
	}

	// Generate UUID if not provided
	if platform.ID == "" {
		platform.ID = uuid.New().String()
	}

	// Set creation timestamp
	if platform.CreatedAt.IsZero() {
		platform.CreatedAt = time.Now().UTC()
	}

	s.platforms = append(s.platforms, platform)
	return nil
}

// ListPlatforms returns platform configurations for a user ordered by name ascending
func (s *MemoryStore) ListPlatforms(userID string) []Platform {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []Platform
	for _, p := range s.platforms {
		if p.UserID == userID {
			out = append(out, p)
		}
	}

	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[i].Name > out[j].Name {
				out[i], out[j] = out[j], out[i]
			}
		}
	}

	return out
}

// AddSubsource stores a new subsource configuration with generated UUID
func (s *MemoryStore) AddSubsource(userID string, subsource Subsource) error {
	// Validate subsource data
	if err := validateSubsource(subsource); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	platformExists := false
	var platformName string
	for _, p := range s.platforms {
		if p.ID == subsource.PlatformID && p.UserID == userID {
			platformExists = true
			platformName = p.Name
			break
		}
	}
	if !platformExists {
		return &ValidationError{Details: []string{fmt.Sprintf("platform not found: %s", subsource.PlatformID)}}
	}
	subsource.UserID = userID

	// Check for duplicate identifier within the same platform
	for _, sub := range s.subsources {
		if sub.PlatformID == subsource.PlatformID && sub.Identifier == subsource.Identifier {
			return &ValidationError{Details: []string{fmt.Sprintf("subsource identifier already exists for platform: %s", subsource.Identifier)}}
		}
	}

	// Generate UUID if not provided
	if subsource.ID == "" {
		subsource.ID = uuid.New().String()
	}

	// Set creation timestamp
	if subsource.CreatedAt.IsZero() {
		subsource.CreatedAt = time.Now().UTC()
	}

	// Auto-generate URL if not provided
	if subsource.URL == "" {
		subsource.URL = generateSubsourceURL(platformName, subsource.Identifier)
	}

	s.subsources = append(s.subsources, subsource)
	return nil
}

// ListAllSubsources returns user-owned subsources (platform has user_id set)
func (s *MemoryStore) ListAllSubsources() []SubsourceWithPlatform {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]SubsourceWithPlatform, 0, len(s.subsources))
	for _, subsource := range s.subsources {
		var platformName string
		var ownerID string
		for _, platform := range s.platforms {
			if platform.ID == subsource.PlatformID {
				platformName = platform.Name
				ownerID = platform.UserID
				break
			}
		}
		if ownerID == "" {
			continue
		}
		subsource.UserID = ownerID
		out = append(out, SubsourceWithPlatform{
			Subsource:    subsource,
			PlatformName: platformName,
		})
	}
	return out
}

// ListSubsources returns subsources for a specific platform with platform information
func (s *MemoryStore) ListSubsources(userID string, platformID string) []SubsourceWithPlatform {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]SubsourceWithPlatform, 0)

	var platformName string
	for _, platform := range s.platforms {
		if platform.ID == platformID && platform.UserID == userID {
			platformName = platform.Name
			break
		}
	}
	if platformName == "" {
		return out
	}

	for _, subsource := range s.subsources {
		if subsource.PlatformID == platformID {
			sub := subsource
			sub.UserID = userID
			out = append(out, SubsourceWithPlatform{
				Subsource:    sub,
				PlatformName: platformName,
			})
		}
	}

	return out
}

// GetSubsource retrieves a subsource by ID with platform information
func (s *MemoryStore) GetSubsource(id string) (SubsourceWithPlatform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subsource := range s.subsources {
		if subsource.ID == id {
			var platformName string
			var ownerID string
			for _, platform := range s.platforms {
				if platform.ID == subsource.PlatformID {
					platformName = platform.Name
					ownerID = platform.UserID
					break
				}
			}
			sub := subsource
			sub.UserID = ownerID
			return SubsourceWithPlatform{
				Subsource:    sub,
				PlatformName: platformName,
			}, true
		}
	}
	return SubsourceWithPlatform{}, false
}

// GetSubsourceForUser returns a subsource only if owned by the user
func (s *MemoryStore) GetSubsourceForUser(userID string, id string) (SubsourceWithPlatform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subsource := range s.subsources {
		if subsource.ID != id {
			continue
		}
		for _, platform := range s.platforms {
			if platform.ID == subsource.PlatformID && platform.UserID == userID {
				sub := subsource
				sub.UserID = userID
				return SubsourceWithPlatform{
					Subsource:    sub,
					PlatformName: platform.Name,
				}, true
			}
		}
	}
	return SubsourceWithPlatform{}, false
}

// UpdateSubsource modifies subsource configuration while preventing platform_id changes
func (s *MemoryStore) UpdateSubsource(userID string, id string, subsource Subsource) error {
	// Validate subsource data
	if err := validateSubsource(subsource); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.subsources {
		if sub.ID != id {
			continue
		}
		if !s.platformOwnedByUser(sub.PlatformID, userID) {
			break
		}
		subsource.CreatedAt = sub.CreatedAt
		subsource.PlatformID = sub.PlatformID
		subsource.UserID = userID
		subsource.ID = id
		s.subsources[i] = subsource
		return nil
	}
	return &ValidationError{Details: []string{fmt.Sprintf("subsource not found: %s", id)}}
}

func (s *MemoryStore) platformOwnedByUser(platformID, userID string) bool {
	for _, p := range s.platforms {
		if p.ID == platformID && p.UserID == userID {
			return true
		}
	}
	return false
}

// DeleteSubsource removes a subsource
func (s *MemoryStore) DeleteSubsource(userID string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.subsources {
		if sub.ID == id && s.platformOwnedByUser(sub.PlatformID, userID) {
			s.subsources = append(s.subsources[:i], s.subsources[i+1:]...)
			return nil
		}
	}
	return nil
}

// GetPlatform retrieves a platform by ID for a user
func (s *MemoryStore) GetPlatform(userID string, id string) (Platform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, platform := range s.platforms {
		if platform.ID == id && platform.UserID == userID {
			return platform, true
		}
	}
	return Platform{}, false
}

// GetPlatformUnscoped retrieves a platform by ID
func (s *MemoryStore) GetPlatformUnscoped(id string) (Platform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, platform := range s.platforms {
		if platform.ID == id {
			return platform, true
		}
	}
	return Platform{}, false
}

// GetPlatformByName retrieves a platform by name for a user
func (s *MemoryStore) GetPlatformByName(userID string, name string) (Platform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, platform := range s.platforms {
		if platform.UserID == userID && platform.Name == name {
			return platform, true
		}
	}
	return Platform{}, false
}

// UpdatePlatform modifies platform configuration while preserving created_at
func (s *MemoryStore) UpdatePlatform(userID string, id string, platform Platform) error {
	normalizePlatformCombines(&platform)
	platform.UserID = userID
	// Validate platform data
	if err := validatePlatform(platform); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.platforms {
		if p.ID == id && p.UserID == userID {
			platform.CreatedAt = p.CreatedAt
			platform.ID = id
			s.platforms[i] = platform
			return nil
		}
	}
	return &ValidationError{Details: []string{fmt.Sprintf("platform not found: %s", id)}}
}

// DeletePlatform removes a platform and cascades to subsources and filters
func (s *MemoryStore) DeletePlatform(userID string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	platformIndex := -1
	for i, p := range s.platforms {
		if p.ID == id && p.UserID == userID {
			platformIndex = i
			break
		}
	}

	if platformIndex == -1 {
		return nil
	}

	// Remove platform
	s.platforms = append(s.platforms[:platformIndex], s.platforms[platformIndex+1:]...)

	// Cascade delete subsources
	newSubsources := make([]Subsource, 0)
	for _, sub := range s.subsources {
		if sub.PlatformID != id {
			newSubsources = append(newSubsources, sub)
		}
	}
	s.subsources = newSubsources

	// Cascade delete filters
	newFilters := make([]DestinationFilter, 0)
	for _, f := range s.filters {
		if f.PlatformID != id {
			newFilters = append(newFilters, f)
		}
	}
	s.filters = newFilters

	return nil
}

func (s *MemoryStore) AddFilter(userID string, filter DestinationFilter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.platformOwnedByUser(filter.PlatformID, userID) {
		return fmt.Errorf("platform not found: %s", filter.PlatformID)
	}
	filter.UserID = userID

	if filter.ID == "" {
		filter.ID = uuid.New().String()
	}
	if filter.CreatedAt.IsZero() {
		filter.CreatedAt = time.Now().UTC()
	}

	s.filters = append(s.filters, filter)
	return nil
}

func (s *MemoryStore) ListFilters(userID string, platformID string) []DestinationFilter {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []DestinationFilter
	if !s.platformOwnedByUser(platformID, userID) {
		return []DestinationFilter{}
	}
	for _, f := range s.filters {
		if f.PlatformID == platformID {
			out = append(out, f)
		}
	}
	if out == nil {
		return []DestinationFilter{}
	}
	return out
}

func (s *MemoryStore) DeleteFilter(userID string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, f := range s.filters {
		if f.ID == id && s.platformOwnedByUser(f.PlatformID, userID) {
			s.filters = append(s.filters[:i], s.filters[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("filter not found: %s", id)
}
