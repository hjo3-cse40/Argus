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
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	DiscordWebhook string    `json:"discord_webhook"`
	WebhookSecret  string    `json:"webhook_secret,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Subsource struct {
	ID         string    `json:"id"`
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

type MemoryStore struct {
	mu         sync.Mutex
	deliveries []Delivery
	sources    []Source
	platforms  []Platform
	subsources []Subsource
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
	if len(s.deliveries) > s.limit {
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

func (s *MemoryStore) List() []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Delivery, len(s.deliveries))
	copy(out, s.deliveries)
	return out
}

func (s *MemoryStore) ListDeliveriesBySubsource(subsourceID string) []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []Delivery
	for _, d := range s.deliveries {
		if d.SubsourceID != nil && *d.SubsourceID == subsourceID {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func (s *MemoryStore) ListDeliveriesByPlatform(platformID string) []Delivery {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build a set of subsource IDs for this platform
	subsourceIDs := make(map[string]bool)
	for _, sub := range s.subsources {
		if sub.PlatformID == platformID {
			subsourceIDs[sub.ID] = true
		}
	}

	// Filter deliveries by subsource IDs
	var filtered []Delivery
	for _, d := range s.deliveries {
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
func (s *MemoryStore) AddPlatform(platform Platform) error {
	// Validate platform data
	if err := validatePlatform(platform); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name
	for _, p := range s.platforms {
		if p.Name == platform.Name {
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

// ListPlatforms returns all platform configurations ordered by name ascending
func (s *MemoryStore) ListPlatforms() []Platform {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Platform, len(s.platforms))
	copy(out, s.platforms)
	
	// Sort by name ascending
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
func (s *MemoryStore) AddSubsource(subsource Subsource) error {
	// Validate subsource data
	if err := validateSubsource(subsource); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if platform exists
	platformExists := false
	var platformName string
	for _, p := range s.platforms {
		if p.ID == subsource.PlatformID {
			platformExists = true
			platformName = p.Name
			break
		}
	}
	if !platformExists {
		return &ValidationError{Details: []string{fmt.Sprintf("platform not found: %s", subsource.PlatformID)}}
	}

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

// ListAllSubsources returns all subsource configurations with platform information
func (s *MemoryStore) ListAllSubsources() []SubsourceWithPlatform {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]SubsourceWithPlatform, 0, len(s.subsources))
	for _, subsource := range s.subsources {
		// Find the platform for this subsource
		var platformName string
		for _, platform := range s.platforms {
			if platform.ID == subsource.PlatformID {
				platformName = platform.Name
				break
			}
		}

		out = append(out, SubsourceWithPlatform{
			Subsource:    subsource,
			PlatformName: platformName,
		})
	}
	return out
}

// ListSubsources returns subsources for a specific platform with platform information
func (s *MemoryStore) ListSubsources(platformID string) []SubsourceWithPlatform {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]SubsourceWithPlatform, 0)
	
	// Find the platform name
	var platformName string
	for _, platform := range s.platforms {
		if platform.ID == platformID {
			platformName = platform.Name
			break
		}
	}

	// Filter subsources by platform_id
	for _, subsource := range s.subsources {
		if subsource.PlatformID == platformID {
			out = append(out, SubsourceWithPlatform{
				Subsource:    subsource,
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
			// Find the platform name
			var platformName string
			for _, platform := range s.platforms {
				if platform.ID == subsource.PlatformID {
					platformName = platform.Name
					break
				}
			}

			return SubsourceWithPlatform{
				Subsource:    subsource,
				PlatformName: platformName,
			}, true
		}
	}
	return SubsourceWithPlatform{}, false
}

// UpdateSubsource modifies subsource configuration while preventing platform_id changes
func (s *MemoryStore) UpdateSubsource(id string, subsource Subsource) error {
	// Validate subsource data
	if err := validateSubsource(subsource); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.subsources {
		if sub.ID == id {
			// Preserve created_at and platform_id (immutable)
			subsource.CreatedAt = sub.CreatedAt
			subsource.PlatformID = sub.PlatformID
			subsource.ID = id
			s.subsources[i] = subsource
			return nil
		}
	}
	return &ValidationError{Details: []string{fmt.Sprintf("subsource not found: %s", id)}}
}

// DeleteSubsource removes a subsource
func (s *MemoryStore) DeleteSubsource(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.subsources {
		if sub.ID == id {
			s.subsources = append(s.subsources[:i], s.subsources[i+1:]...)
			return nil
		}
	}
	return nil
}


// GetPlatform retrieves a platform by ID
func (s *MemoryStore) GetPlatform(id string) (Platform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, platform := range s.platforms {
		if platform.ID == id {
			return platform, true
		}
	}
	return Platform{}, false
}

// GetPlatformByName retrieves a platform by name
func (s *MemoryStore) GetPlatformByName(name string) (Platform, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, platform := range s.platforms {
		if platform.Name == name {
			return platform, true
		}
	}
	return Platform{}, false
}

// UpdatePlatform modifies platform configuration while preserving created_at
func (s *MemoryStore) UpdatePlatform(id string, platform Platform) error {
	// Validate platform data
	if err := validatePlatform(platform); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.platforms {
		if p.ID == id {
			// Preserve created_at
			platform.CreatedAt = p.CreatedAt
			platform.ID = id
			s.platforms[i] = platform
			return nil
		}
	}
	return &ValidationError{Details: []string{fmt.Sprintf("platform not found: %s", id)}}
}

// DeletePlatform removes a platform and cascades to subsources
func (s *MemoryStore) DeletePlatform(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove platform
	platformIndex := -1
	for i, p := range s.platforms {
		if p.ID == id {
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

	return nil

}
