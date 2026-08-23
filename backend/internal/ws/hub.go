package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/model"
	"gopomodoro/internal/pomodoro"
)

type client struct {
	id           string
	userID       uuid.UUID
	send         chan Outbound
	milestones   map[uuid.UUID]struct{}
	sessionWatch uuid.UUID
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*client
	byUser  map[uuid.UUID]map[string]*client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*client),
		byUser:  make(map[uuid.UUID]map[string]*client),
	}
}

func (h *Hub) Add(userID uuid.UUID) *client {
	c := &client{
		id:         uuid.NewString(),
		userID:     userID,
		send:       make(chan Outbound, 32),
		milestones: make(map[uuid.UUID]struct{}),
	}
	h.mu.Lock()
	h.clients[c.id] = c
	if h.byUser[userID] == nil {
		h.byUser[userID] = make(map[string]*client)
	}
	h.byUser[userID][c.id] = c
	h.mu.Unlock()
	return c
}

func (h *Hub) Remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c.id)
	if m := h.byUser[c.userID]; m != nil {
		delete(m, c.id)
		if len(m) == 0 {
			delete(h.byUser, c.userID)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Subscribe(c *client, mid uuid.UUID) {
	h.mu.Lock()
	c.milestones[mid] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) BroadcastSession(userID model.ID, view pomodoro.SessionView) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.byUser[userID] {
		h.trySend(c, Outbound{Type: TypeSession, Payload: view})
		h.trySend(c, Outbound{Type: TypeTick, Payload: map[string]any{
			"session_id":   view.Session.ID,
			"state":        view.Session.State,
			"remaining_ms": view.RemainingMS,
			"server_now":   time.Now().Format(time.RFC3339),
		}})
	}
}

func (h *Hub) BroadcastGrace(userID model.ID, leftS int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.byUser[userID] {
		h.trySend(c, Outbound{Type: TypeGrace, Payload: map[string]any{"remaining_s": leftS}})
	}
}

func (h *Hub) BroadcastBurndown(milestoneID model.ID, point model.BurndownPoint) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if _, ok := c.milestones[milestoneID]; ok {
			h.trySend(c, Outbound{Type: TypeBurndown, Payload: point})
		}
	}
}

func (h *Hub) trySend(c *client, msg Outbound) {
	select {
	case c.send <- msg:
	default:
	}
}
