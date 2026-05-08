package store

// Store defines the interface for storage operations
type Store interface {
	// Delivery operations
	AddQueued(d Delivery)
	MarkDelivered(eventID string) bool
	MarkFailed(eventID string, retryCount int, lastErr string) bool
	GetDelivery(eventID string) (Delivery, bool)
	List(userID string) []Delivery
	ListDeliveriesBySubsource(userID string, subsourceID string) []Delivery
	ListDeliveriesByPlatform(userID string, platformID string) []Delivery

	// Source operations
	AddSource(source Source) error
	ListSources() []Source
	GetSource(id string) (Source, bool)
	GetSourceByName(name string) (Source, bool)

	// Platform operations (scoped to user for API privacy)
	AddPlatform(userID string, platform Platform) error
	ListPlatforms(userID string) []Platform
	GetPlatform(userID string, id string) (Platform, bool)
	GetPlatformByName(userID string, name string) (Platform, bool)
	UpdatePlatform(userID string, id string, platform Platform) error
	DeletePlatform(userID string, id string) error

	// GetPlatformUnscoped resolves by id only (worker / internal; IDs are UUIDs)
	GetPlatformUnscoped(id string) (Platform, bool)

	// Subsource operations
	AddSubsource(userID string, subsource Subsource) error
	ListSubsources(userID string, platformID string) []SubsourceWithPlatform
	ListAllSubsources() []SubsourceWithPlatform // user-owned only (platform.user_id NOT NULL)
	GetSubsource(id string) (SubsourceWithPlatform, bool) // unscoped (worker)
	GetSubsourceForUser(userID string, id string) (SubsourceWithPlatform, bool)
	UpdateSubsource(userID string, id string, subsource Subsource) error
	DeleteSubsource(userID string, id string) error

	// Filter operations
	AddFilter(userID string, filter DestinationFilter) error
	ListFilters(userID string, platformID string) []DestinationFilter
	DeleteFilter(userID string, id string) error

	// Auth Operations
	CreateUser(user User) error
	GetUserByEmail(email string) (User, bool)
	GetUserByID(id string) (User, bool)

	CreateSession(session Session) error
	GetSession(sessionID string) (Session, bool)
	DeleteSession(sessionID string) error
	DeleteExpiredSessions() error
	// Lifecycle
	Close() error
}
