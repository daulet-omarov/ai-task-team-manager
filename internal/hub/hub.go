package hub

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

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type client struct {
	hub     *Hub
	conn    *websocket.Conn
	boardID uint
	send    chan []byte
}

func (c *client) writePump() {
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

func (c *client) readPump() {
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
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// Hub manages WebSocket connections grouped by boardID.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[uint]map[*client]struct{}
	register   chan *client
	unregister chan *client
}

func New() *Hub {
	h := &Hub{
		rooms:      make(map[uint]map[*client]struct{}),
		register:   make(chan *client, 64),
		unregister: make(chan *client, 64),
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
				h.rooms[c.boardID] = make(map[*client]struct{})
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

// Broadcast sends an Event to all clients connected to the given board room.
func (h *Hub) Broadcast(boardID uint, event Event) {
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
		}
	}
}

// Connect upgrades conn to a WebSocket client in the given board room and
// blocks until the connection is closed.
func (h *Hub) Connect(conn *websocket.Conn, boardID uint) {
	c := &client{
		hub:     h,
		conn:    conn,
		boardID: boardID,
		send:    make(chan []byte, 256),
	}
	h.register <- c
	go c.writePump()
	c.readPump()
}
