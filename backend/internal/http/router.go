package http

import (
	"log"
	"net/http"
	"os"
	"strings"

	"argus-backend/internal/auth"
	"argus-backend/internal/http/handlers"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

func NewRouter(mqClient *mq.Client, st store.Store, authService *auth.Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)

	debug := handlers.NewDebugPublisher(mqClient, st)
	mux.HandleFunc("POST /debug/publish", debug.Publish)

	markQueued := handlers.NewMarkQueuedHandler(st)
	mux.HandleFunc("POST /debug/queued", markQueued.MarkQueued)

	dh := handlers.NewDeliveriesHandler(st)
	mux.HandleFunc("GET /deliveries", dh.List)

	mark := handlers.NewMarkDeliveredHandler(st)
	mux.HandleFunc("POST /debug/delivered", mark.Mark)

	markFailed := handlers.NewMarkFailedHandler(st)
	mux.HandleFunc("POST /debug/failed", markFailed.Mark)

	// Ingestion: normalize and enqueue events
	ingest := handlers.NewIngestHandler(mqClient, st)
	mux.HandleFunc("POST /api/ingest", ingest.Ingest)

	// Source management endpoints
	sh := handlers.NewSourcesHandler(st)
	mux.HandleFunc("POST /api/sources", sh.Create)
	mux.HandleFunc("GET /api/sources", sh.List)

	// Platform management endpoints
	ph := handlers.NewPlatformsHandler(st)
	mux.HandleFunc("POST /api/platforms", ph.Create)
	mux.HandleFunc("GET /api/platforms", ph.List)
	mux.HandleFunc("GET /api/platforms/{id}", ph.Get)
	mux.HandleFunc("PUT /api/platforms/{id}", ph.Update)
	mux.HandleFunc("DELETE /api/platforms/{id}", ph.Delete)

	// Subsource management endpoints
	subh := handlers.NewSubsourcesHandler(st)
	mux.HandleFunc("POST /api/platforms/{platform_id}/subsources", subh.Create)
	mux.HandleFunc("GET /api/platforms/{platform_id}/subsources", subh.ListByPlatform)
	mux.HandleFunc("GET /api/subsources/{id}", subh.Get)
	mux.HandleFunc("PUT /api/subsources/{id}", subh.Update)
	mux.HandleFunc("DELETE /api/subsources/{id}", subh.Delete)

	// Filter management endpoints
	fh := handlers.NewFiltersHandler(st)
	mux.HandleFunc("POST /api/platforms/{platform_id}/filters", fh.Create)
	mux.HandleFunc("GET /api/platforms/{platform_id}/filters", fh.List)
	mux.HandleFunc("DELETE /api/filters/{id}", fh.Delete)

	// Auth endpoints
	ah := handlers.NewAuthHandler(authService)
	mux.HandleFunc("POST /api/auth/register", ah.Register)
	mux.HandleFunc("POST /api/auth/login", ah.Login)
	mux.HandleFunc("POST /api/auth/logout", ah.Logout)
	mux.HandleFunc("GET /api/auth/me", authService.RequireAuth(http.HandlerFunc(ah.Me)).ServeHTTP)

	// Test endpoint to serve CSS directly
	mux.HandleFunc("GET /test-css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "../static/css/styles.css")
	})

	// Serve static files for the UI
	// Try different paths to find the static directory
	var staticDir string
	if _, err := os.Stat("static"); err == nil {
		staticDir = "static" // Running from project root
	} else if _, err := os.Stat("../static"); err == nil {
		staticDir = "../static" // Running from backend directory
	} else {
		log.Fatal("Could not find static directory")
	}

	log.Printf("Serving static files from: %s", staticDir)

	// Serve CSS files - MUST be registered before catch-all handler
	mux.HandleFunc("GET /css/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		log.Printf("Serving CSS file: %s from %s", r.URL.Path, staticDir)
		http.ServeFile(w, r, staticDir+"/css/styles.css")
	})

	// Also handle CSS with query parameters (cache busting)
	mux.HandleFunc("/css/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		log.Printf("Serving CSS file with params: %s from %s", r.URL.String(), staticDir)
		http.ServeFile(w, r, staticDir+"/css/styles.css")
	})

	// Serve JS files - MUST be registered before catch-all handler
	mux.HandleFunc("GET /js/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		log.Printf("Serving JS file: %s from %s", r.URL.Path, staticDir)
		http.ServeFile(w, r, staticDir+"/js/app.js")
	})

	// Also handle JS with query parameters (cache busting)
	mux.HandleFunc("/js/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		log.Printf("Serving JS file with params: %s from %s", r.URL.String(), staticDir)
		http.ServeFile(w, r, staticDir+"/js/app.js")
	})
	// Serve auth.js - handles GET /js/auth.js
	mux.HandleFunc("GET /js/auth.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, staticDir+"/js/auth.js")
	})

	// Also handle auth.js with query parameters (cache busting)
	mux.HandleFunc("/js/auth.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, staticDir+"/js/auth.js")
	})

	// Serve the main HTML file and other static assets - MUST be last
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Don't serve API routes through static handler
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/debug/") || strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/deliveries") || strings.HasPrefix(r.URL.Path, "/test-") {
			http.NotFound(w, r)
			return
		}

		// Don't serve CSS/JS through catch-all - they have specific handlers
		if strings.HasPrefix(r.URL.Path, "/css/") || strings.HasPrefix(r.URL.Path, "/js/") {
			http.NotFound(w, r)
			return
		}

		// For root path, serve landing page
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/index.html", http.StatusFound)
			return
		}

		// Serve index.html (landing page) without redirect loop
		if r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html")
			f, err := os.Open(staticDir + "/index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			defer func() { _ = f.Close() }()
			fi, _ := f.Stat()
			http.ServeContent(w, r, "index.html", fi.ModTime(), f)
			return
		}

		// For other paths, try to serve the file
		log.Printf("Serving static file: %s from %s", r.URL.Path, staticDir)
		http.ServeFile(w, r, staticDir+r.URL.Path)
	})

	return mux
}
