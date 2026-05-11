package realtime

import (
	"context"
	"encoding/json"
	"sync"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
	broadcast   chan []byte
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[chan []byte]struct{}),
		broadcast:   make(chan []byte, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case msg := <-h.broadcast:
			h.mu.RLock()
			for ch := range h.subscribers {
				select {
				case ch <- msg:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, 64)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.subscribers[ch]; ok {
		delete(h.subscribers, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *Hub) Publish(event Event) {
	if h == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	select {
	case h.broadcast <- payload:
	default:
	}
}

func (h *Hub) PublishRaw(message []byte) {
	if h == nil {
		return
	}
	select {
	case h.broadcast <- message:
	default:
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		close(ch)
		delete(h.subscribers, ch)
	}
}
