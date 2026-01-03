package application

import (
	"editor/domain"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected user session.
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// EditorHub orchestrates all active connections and synchronizes document state.
type EditorHub struct {
	Document   *domain.Document
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan domain.Delta

	mu sync.RWMutex
}

// NewEditorHub initializes a hub for a specific document.
func NewEditorHub(document *domain.Document) *EditorHub {
	return &EditorHub{
		Document:   document,
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan domain.Delta),
	}
}

// Run starts the main event loop for the hub, processing registrations, departures
// and document updates sequentially to ensure consistency.
func (h *EditorHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()

			log.Printf("Client %s joined", client.ID)

			h.Document.Mutex.RLock()
			currentState := h.Document.Content
			version := h.Document.Version
			h.Document.Mutex.RUnlock()

			initPayload, _ := json.Marshal(map[string]any{
				"version":   version,
				"full_text": currentState,
			})

			// Send only to the new client
			client.Send <- initPayload
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()

			log.Printf("Client %s left", client.ID)
		case delta := <-h.Broadcast:
			// update the domain state
			newText := h.Document.Transition(&delta)

			// prepare the payload
			message, _ := json.Marshal(map[string]any{
				"version":   h.Document.Version,
				"full_text": newText,
			})

			// broadcast to all clients
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}
