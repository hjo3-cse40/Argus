package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"argus-backend/internal/store"
)

type DeliveryBroadcaster struct {
	mu       sync.Mutex
	nextID   int
	clients  map[int]chan store.Delivery
	closed   bool
}

func NewDeliveryBroadcaster() *DeliveryBroadcaster {
	return &DeliveryBroadcaster{
		clients: make(map[int]chan store.Delivery),
	}
}

func (b *DeliveryBroadcaster) Subscribe() (int, chan store.Delivery) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, nil
	}

	b.nextID++
	id := b.nextID
	ch := make(chan store.Delivery, 16)
	b.clients[id] = ch
	return id, ch
}

func (b *DeliveryBroadcaster) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch, ok := b.clients[id]
	if !ok {
		return
	}
	delete(b.clients, id)
	close(ch)
}

func (b *DeliveryBroadcaster) Publish(d store.Delivery) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	for _, ch := range b.clients {
		select {
		case ch <- d:
		default:
			// Drop when a slow client cannot keep up.
		}
	}
}

type DeliveriesStreamHandler struct {
	Broadcaster *DeliveryBroadcaster
}

func NewDeliveriesStreamHandler(b *DeliveryBroadcaster) *DeliveriesStreamHandler {
	return &DeliveriesStreamHandler{Broadcaster: b}
}

func (h *DeliveriesStreamHandler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	clientID, events := h.Broadcaster.Subscribe()
	if events == nil {
		http.Error(w, "stream unavailable", http.StatusServiceUnavailable)
		return
	}
	defer h.Broadcaster.Unsubscribe(clientID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	_, _ = fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case delivery := <-events:
			payload, err := json.Marshal(delivery)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: delivered\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
