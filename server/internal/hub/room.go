package hub

import (
	"sync"

	"github.com/bullshit-wtf/server/internal/game"
)

// Room manages all WebSocket clients for a single game.
type Room struct {
	Game    *game.Game
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewRoom(g *game.Game) *Room {
	return &Room{
		Game:    g,
		clients: make(map[string]*Client),
	}
}

// AddClient registers a client to this room.
func (r *Room) AddClient(uuid string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[uuid] = client
}

// RemoveClient removes a client from this room.
func (r *Room) RemoveClient(uuid string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, uuid)
}

// GetClient returns the client for a given UUID.
func (r *Room) GetClient(uuid string) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.clients[uuid]
}

// Broadcast sends a message to all clients in the room.
func (r *Room) Broadcast(data []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.clients {
		c.Send(data)
	}
}

// BroadcastExcept sends a message to all clients except the given UUID.
func (r *Room) BroadcastExcept(data []byte, exceptUUID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for uuid, c := range r.clients {
		if uuid != exceptUUID {
			c.Send(data)
		}
	}
}

// ClientCount returns the number of connected clients.
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// IsEmpty returns true if no clients are connected.
func (r *Room) IsEmpty() bool {
	return r.ClientCount() == 0
}

// ForEachClient calls fn for each connected client.
func (r *Room) ForEachClient(fn func(uuid string, client *Client)) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for uuid, c := range r.clients {
		fn(uuid, c)
	}
}
