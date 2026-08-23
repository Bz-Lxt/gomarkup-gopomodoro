package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/logger"
	"gopomodoro/internal/pomodoro"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	Hub      *Hub
	Registry *pomodoro.Registry
	Ping     time.Duration
	PongWait time.Duration
}

func (h *Handler) Serve(c *gin.Context) {
	userID := auth.UserID(c)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.L().Error("ws upgrade failed", "err", err)
		return
	}
	client := h.Hub.Add(userID)
	defer func() {
		h.Hub.Remove(client)
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(h.PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(h.PongWait))
	})

	done := make(chan struct{})
	go h.writeLoop(conn, client, done)

	if view := h.Registry.Active(userID); view != nil {
		h.Registry.MarkConnected(view.Session.ID)
		client.send <- Outbound{Type: TypeSession, Payload: view}
		defer h.Registry.MarkDisconnected(view.Session.ID)
	}

	for {
		var in Inbound
		if err := conn.ReadJSON(&in); err != nil {
			close(done)
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(h.PongWait))
		switch in.Type {
		case TypePing:
			select {
			case client.send <- Outbound{Type: TypePong}:
			default:
			}
		case TypeHello:
			view, err := h.Registry.ResumeByToken(c.Request.Context(), userID, in.ResumeToken)
			if err != nil {
				client.send <- Outbound{Type: TypeError, Payload: httpx.ErrNotFound.Message}
				continue
			}
			if view != nil {
				h.Registry.MarkConnected(view.Session.ID)
				client.send <- Outbound{Type: TypeSession, Payload: view}
			}
		case TypeSubscribe:
			mid, err := uuid.Parse(in.MilestoneID)
			if err != nil {
				continue
			}
			h.Hub.Subscribe(client, mid)
		}
	}
}

func (h *Handler) writeLoop(conn *websocket.Conn, c *client, done <-chan struct{}) {
	ticker := time.NewTicker(h.Ping)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}
