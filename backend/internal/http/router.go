package http

import (
	"net/http"

	"argus-backend/internal/http/handlers"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
)

func NewRouter(mqClient *mq.Client, st *store.MemoryStore) http.Handler {
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

	// Ingestion: normalize and enqueue events
	ingest := handlers.NewIngestHandler(mqClient, st)
	mux.HandleFunc("POST /api/ingest", ingest.Ingest)

	// Source management endpoints
	sh := handlers.NewSourcesHandler(st)
	mux.HandleFunc("POST /api/sources", sh.Create)
	mux.HandleFunc("GET /api/sources", sh.List)

	// Serve static files for the UI (from parent directory)
	fs := http.FileServer(http.Dir("../static"))
	mux.Handle("/", fs)

	return mux
}


