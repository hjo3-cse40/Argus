package http

import (
	"net/http"

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
	require := func(next http.HandlerFunc) http.Handler {
		return authService.RequireAuth(next)
	}

	mux.Handle("GET /deliveries", require(dh.List))
	mux.Handle("GET /api/deliveries", require(dh.List))
	broadcaster := handlers.NewDeliveryBroadcaster()
	stream := handlers.NewDeliveriesStreamHandler(broadcaster)
	mux.Handle("GET /api/deliveries/stream", require(stream.Stream))

	mark := handlers.NewMarkDeliveredHandler(st, broadcaster)
	mux.HandleFunc("POST /debug/delivered", mark.Mark)

	markFailed := handlers.NewMarkFailedHandler(st)
	mux.HandleFunc("POST /debug/failed", markFailed.Mark)

	// Ingestion: normalize and enqueue events
	ingest := handlers.NewIngestHandler(mqClient, st)
	mux.HandleFunc("POST /api/ingest", ingest.Ingest)

	// Source management endpoints
	sh := handlers.NewSourcesHandler(st)
	mux.Handle("POST /api/sources", require(sh.Create))
	mux.Handle("GET /api/sources", require(sh.List))

	// Platform management endpoints
	ph := handlers.NewPlatformsHandler(st)
	mux.Handle("POST /api/platforms", require(ph.Create))
	mux.Handle("GET /api/platforms", require(ph.List))
	mux.Handle("GET /api/platforms/{id}", require(ph.Get))
	mux.Handle("PUT /api/platforms/{id}", require(ph.Update))
	mux.Handle("DELETE /api/platforms/{id}", require(ph.Delete))

	// Subsource management endpoints
	subh := handlers.NewSubsourcesHandler(st)
	mux.Handle("POST /api/platforms/{platform_id}/subsources", require(subh.Create))
	mux.Handle("GET /api/platforms/{platform_id}/subsources", require(subh.ListByPlatform))
	mux.Handle("GET /api/subsources/{id}", require(subh.Get))
	mux.Handle("PUT /api/subsources/{id}", require(subh.Update))
	mux.Handle("DELETE /api/subsources/{id}", require(subh.Delete))

	// Filter management endpoints
	fh := handlers.NewFiltersHandler(st)
	mux.Handle("POST /api/platforms/{platform_id}/filters", require(fh.Create))
	mux.Handle("GET /api/platforms/{platform_id}/filters", require(fh.List))
	mux.Handle("DELETE /api/filters/{id}", require(fh.Delete))

	// Auth endpoints
	ah := handlers.NewAuthHandler(authService)
	mux.HandleFunc("POST /api/auth/register", ah.Register)
	mux.HandleFunc("POST /api/auth/login", ah.Login)
	mux.HandleFunc("POST /api/auth/logout", ah.Logout)
	mux.HandleFunc("GET /api/auth/me", authService.RequireAuth(http.HandlerFunc(ah.Me)).ServeHTTP)

	return mux
}
