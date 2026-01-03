package infrastructure

import (
	"editor/application"
	"editor/domain"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSServer struct {
	hub *application.EditorHub
}

func NewWSServer(hub *application.EditorHub) *WSServer {
	return &WSServer{hub: hub}
}

func (h *WSServer) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading: %v", err)
		return
	}

	client := &application.Client{ID: uuid.New().String(), Conn: conn, Send: make(chan []byte, 512)}

	defer func() {
		h.hub.Unregister <- client
		conn.Close()
	}()

	h.hub.Register <- client

	// Start the Write Loop (Server -> Client).
	go func() {
		for {
			message, ok := <-client.Send
			if !ok {
				// Hub closed the channel
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}()

	// Start the Read Loop (Client -> Server).
	for {
		var delta domain.Delta
		err := conn.ReadJSON(&delta)
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}
		delta.UserID = client.ID
		h.hub.Broadcast <- delta
	}
}
