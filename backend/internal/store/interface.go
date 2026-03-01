package store

// Store defines the interface for storage operations
type Store interface {
	// Delivery operations
	AddQueued(d Delivery)
	MarkDelivered(eventID string) bool
	MarkFailed(eventID string, retryCount int, lastErr string) bool
	List() []Delivery
	ListDeliveriesBySubsource(subsourceID string) []Delivery
	ListDeliveriesByPlatform(platformID string) []Delivery

	// Source operations
	AddSource(source Source) error
	ListSources() []Source
	GetSource(id string) (Source, bool)
	GetSourceByName(name string) (Source, bool)

	// Platform operations
	AddPlatform(platform Platform) error
	ListPlatforms() []Platform
	GetPlatform(id string) (Platform, bool)
	GetPlatformByName(name string) (Platform, bool)
	UpdatePlatform(id string, platform Platform) error
	DeletePlatform(id string) error

	// Subsource operations
	AddSubsource(subsource Subsource) error
	ListSubsources(platformID string) []SubsourceWithPlatform
	ListAllSubsources() []SubsourceWithPlatform
	GetSubsource(id string) (SubsourceWithPlatform, bool)
	UpdateSubsource(id string, subsource Subsource) error
	DeleteSubsource(id string) error

	// Lifecycle
	Close() error
}
