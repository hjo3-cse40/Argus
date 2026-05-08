package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"argus-backend/internal/auth"
	"argus-backend/internal/http/handlers"
	"argus-backend/internal/store"
)

const frontendTestUserID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func main() {
	log.Println("Starting frontend test server (no RabbitMQ required)")

	// Use in-memory store for testing
	st := store.NewMemoryStore(50)
	
	// Create a simple router without MQ dependency
	mux := http.NewServeMux()

	// Add some test data
	addTestData(st)

	withUser := func(h http.HandlerFunc) http.HandlerFunc {
		u := store.User{ID: frontendTestUserID, Email: "frontend-test@local"}
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r.WithContext(auth.ContextWithUser(r.Context(), u)))
		}
	}

	// Platform management endpoints
	ph := handlers.NewPlatformsHandler(st)
	mux.HandleFunc("POST /api/platforms", withUser(ph.Create))
	mux.HandleFunc("GET /api/platforms", withUser(ph.List))
	mux.HandleFunc("GET /api/platforms/{id}", withUser(ph.Get))
	mux.HandleFunc("PUT /api/platforms/{id}", withUser(ph.Update))
	mux.HandleFunc("DELETE /api/platforms/{id}", withUser(ph.Delete))

	// Subsource management endpoints
	sh := handlers.NewSubsourcesHandler(st)
	mux.HandleFunc("POST /api/platforms/{platform_id}/subsources", withUser(sh.Create))
	mux.HandleFunc("GET /api/platforms/{platform_id}/subsources", withUser(sh.ListByPlatform))
	mux.HandleFunc("GET /api/subsources/{id}", withUser(sh.Get))
	mux.HandleFunc("PUT /api/subsources/{id}", withUser(sh.Update))
	mux.HandleFunc("DELETE /api/subsources/{id}", withUser(sh.Delete))

	// Serve static files - handle both from backend dir and root dir
	var staticDir string
	if _, err := os.Stat("static"); err == nil {
		staticDir = "static"  // Running from project root
	} else if _, err := os.Stat("../static"); err == nil {
		staticDir = "../static"  // Running from backend directory
	} else {
		log.Fatal("Could not find static directory")
	}
	
	log.Printf("Serving static files from: %s", staticDir)
	
	// Serve static files for everything that's not an API route
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Frontend test server listening on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func addTestData(st *store.MemoryStore) {
	_ = st.CreateUser(store.User{
		ID:           frontendTestUserID,
		Email:        "frontend-test@local",
		PasswordHash: "-",
		CreatedAt:    time.Now().UTC(),
	})
	youtube := store.Platform{
		Name:           "youtube",
		DiscordWebhook: "https://discord.com/api/webhooks/123/test",
		WebhookSecret:  "secret123",
	}
	_ = st.AddPlatform(frontendTestUserID, youtube)

	reddit := store.Platform{
		Name:           "reddit",
		DiscordWebhook: "https://discord.com/api/webhooks/456/test",
		WebhookSecret:  "secret456",
	}
	_ = st.AddPlatform(frontendTestUserID, reddit)

	log.Println("Added test platforms: youtube, reddit")
}