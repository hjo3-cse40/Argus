package store

import (
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
	EventID    string        `json:"event_id"`
	Source     string        `json:"source"`
	Title      string        `json:"title"`
	URL        string        `json:"url"`
	Status     DeliveryStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	RetryCount int			 `json:"retry_count"`
	LastError  string		 `json:"last_error,omitempty"`
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

type MemoryStore struct {
	mu         sync.Mutex
	deliveries []Delivery
	sources    []Source
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
