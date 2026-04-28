package chat

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

type WSEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type wsClient struct {
	hub     *Hub
	conn    *websocket.Conn
	boardID uint
	send    chan []byte
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump keeps the connection alive and handles client disconnects.
func (c *wsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		// We don't expect messages from clients, just drain to keep alive.
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// Hub manages all active WebSocket connections, grouped by board.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[uint]map[*wsClient]struct{}
	register   chan *wsClient
	unregister chan *wsClient
}

func NewHub() *Hub {
	h := &Hub{
		rooms:      make(map[uint]map[*wsClient]struct{}),
		register:   make(chan *wsClient, 64),
		unregister: make(chan *wsClient, 64),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.boardID] == nil {
				h.rooms[c.boardID] = make(map[*wsClient]struct{})
			}
			h.rooms[c.boardID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[c.boardID]; ok {
				if _, exists := room[c]; exists {
					delete(room, c)
					close(c.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a WSEvent to all clients connected to the given board.
func (h *Hub) Broadcast(boardID uint, event WSEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	room := h.rooms[boardID]
	h.mu.RUnlock()
	for c := range room {
		select {
		case c.send <- data:
		default:
			// slow client — drop the message rather than blocking
		}
	}
}
