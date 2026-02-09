package events

import (
	"encoding/json"
	"time"
)

// Event represents a raw event in the system
type Event struct {
	EventID   string    `json:"event_id"`
	Source    string    `json:"source"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	// Optional fields for future expansion
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validate checks if the event has required fields
func (e *Event) Validate() error {
	if e.EventID == "" {
		return ErrMissingEventID
	}
	if e.Source == "" {
		return ErrMissingSource
	}
	if e.Title == "" {
		return ErrMissingTitle
	}
	if e.URL == "" {
		return ErrMissingURL
	}
	return nil
}

// ToJSON converts the event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON bytes
func FromJSON(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// NewEvent creates a new event with the current timestamp
func NewEvent(eventID, source, title, url string) *Event {
	return &Event{
		EventID:   eventID,
		Source:    source,
		Title:     title,
		URL:       url,
		CreatedAt: time.Now().UTC(),
		Metadata:  make(map[string]interface{}),
	}
}
