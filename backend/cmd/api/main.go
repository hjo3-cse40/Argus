package main

import (
	"log"
	"net/http"
	"time"

	apphttp "argus-backend/internal/http"
	"argus-backend/internal/config"
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
	st := store.NewMemoryStore(50)
	handler := apphttp.NewRouter(mqClient, st)

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


