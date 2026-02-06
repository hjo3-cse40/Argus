package store

import (
	"sync"
	"time"
)

type DeliveryStatus string

const (
	StatusQueued    DeliveryStatus = "queued"
	StatusDelivered DeliveryStatus = "delivered"
)

type Delivery struct {
	EventID    string        `json:"event_id"`
	Source     string        `json:"source"`
	Title      string        `json:"title"`
	URL        string        `json:"url"`
	Status     DeliveryStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type MemoryStore struct {
	mu        sync.Mutex
	deliveries []Delivery
	limit     int
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
