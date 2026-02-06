package main

import (
	"log"
	"net/http"
	"os"
	"time"

	apphttp "argus-backend/internal/http"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://argus:argus@localhost:5672/"
	}

	mqClient, err := mq.Connect(amqpURL)
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
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}


