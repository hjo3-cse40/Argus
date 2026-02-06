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

	dh := handlers.NewDeliveriesHandler(st)
	mux.HandleFunc("GET /deliveries", dh.List)

	mark := handlers.NewMarkDeliveredHandler(st)
	mux.HandleFunc("POST /debug/delivered", mark.Mark)

	return mux
}


