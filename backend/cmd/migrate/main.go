package main

import (
	"log"
	"os"

	"argus-backend/internal/config"
	"argus-backend/internal/store"
)

func main() {
	log.Println("Starting database migration...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("ERROR: Failed to load configuration: %v", err)
		os.Exit(1)
	}

	log.Printf("Connecting to database at %s:%s...", cfg.Database.Host, cfg.Database.Port)

	// Initialize store (which automatically runs migrations)
	st, err := store.NewPostgresStore(cfg.Database.ConnectionString(), cfg.DeliveryLimit)
	if err != nil {
		log.Printf("ERROR: Migration failed: %v", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	log.Println("SUCCESS: Database migrations completed successfully")
	os.Exit(0)
}
