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

// clientsForUser returns a snapshot of the clients subscribed to a user's
// session updates. The snapshot is collected under the read lock so callers
// may deliver messages without holding it; this is required because
// connection removal (Remove) mutates the underlying maps concurrently with
// broadcasts. Iterating the live map after releasing the lock is what
// previously triggered "concurrent map iteration and map write".
func (h *Hub) clientsForUser(userID model.ID) []*client {
	h.mu.RLock()
	src := h.byUser[userID]
	out := make([]*client, 0, len(src))
	for _, c := range src {
		out = append(out, c)
	}
	h.mu.RUnlock()
	return out
}

// subscribers returns a snapshot of clients watching the given milestone.
func (h *Hub) subscribers(milestoneID model.ID) []*client {
	h.mu.RLock()
	out := make([]*client, 0, len(h.clients))
	for _, c := range h.clients {
		if _, ok := c.milestones[milestoneID]; ok {
			out = append(out, c)
		}
	}
	h.mu.RUnlock()
	return out
}

func (h *Hub) BroadcastSession(userID model.ID, view pomodoro.SessionView) {
	tick := Outbound{Type: TypeTick, Payload: map[string]any{
		"session_id":   view.Session.ID,
		"state":        view.Session.State,
		"remaining_ms": view.RemainingMS,
		"server_now":   time.Now().Format(time.RFC3339),
	}}
	msg := Outbound{Type: TypeSession, Payload: view}
	for _, c := range h.clientsForUser(userID) {
		h.trySend(c, msg)
		h.trySend(c, tick)
	}
}

func (h *Hub) BroadcastGrace(userID model.ID, leftS int) {
	msg := Outbound{Type: TypeGrace, Payload: map[string]any{"remaining_s": leftS}}
	for _, c := range h.clientsForUser(userID) {
		h.trySend(c, msg)
	}
}

func (h *Hub) BroadcastBurndown(milestoneID model.ID, point model.BurndownPoint) {
	msg := Outbound{Type: TypeBurndown, Payload: point}
	for _, c := range h.subscribers(milestoneID) {
		h.trySend(c, msg)
	}
}

func (h *Hub) trySend(c *client, msg Outbound) {
	select {
	case c.send <- msg:
	default:
	}
}
