package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type AlertMessage struct {
	Type      string         `json:"type"`
	Severity  string         `json:"severity"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

type Client struct {
	UserID string
	Role   string
	Conn   *websocket.Conn
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	incoming   chan routedAlert
}

type routedAlert struct {
	msg      AlertMessage
	roles    []string
	userIDs  []string
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		incoming:   make(chan routedAlert, 256),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, c)
			h.mu.Unlock()
		case alert := <-h.incoming:
			payload, err := json.Marshal(alert.msg)
			if err != nil {
				continue
			}
			h.mu.RLock()
			for c := range h.clients {
				if !shouldDeliver(c, alert.roles, alert.userIDs) {
					continue
				}
				_ = c.Conn.WriteMessage(websocket.TextMessage, payload)
			}
			h.mu.RUnlock()
		}
	}
}

func shouldDeliver(c *Client, roles, userIDs []string) bool {
	if len(roles) == 0 && len(userIDs) == 0 {
		return true
	}
	for _, id := range userIDs {
		if id == c.UserID {
			return true
		}
	}
	for _, role := range roles {
		if role == c.Role {
			return true
		}
	}
	return false
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Publish(msg AlertMessage, roles, userIDs []string) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	h.incoming <- routedAlert{msg: msg, roles: roles, userIDs: userIDs}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
