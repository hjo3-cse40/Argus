package main

import (
	"log"
	"net/http"
	"time"

	"argus-backend/internal/auth"
	"argus-backend/internal/config"
	apphttp "argus-backend/internal/http"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Starting API in %s environment", cfg.Environment)

	mqClient, err := mq.Connect(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer mqClient.Close()

	if err := mqClient.DeclareQueue("raw_events"); err != nil {
		log.Fatal(err)
	}

	// Initialize PostgreSQL store
	st, err := store.NewPostgresStore(cfg.Database.ConnectionString(), cfg.DeliveryLimit)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer st.Close()

	log.Printf("Connected to PostgreSQL database at %s:%s", cfg.Database.Host, cfg.Database.Port)

	authService := auth.NewService(st)
	handler := apphttp.NewRouter(mqClient, st, authService)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
