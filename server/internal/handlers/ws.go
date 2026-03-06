package handlers

import (
	"log"
	"net/http"

	"github.com/bullshit-wtf/server/internal/hub"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func wsHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}

		client := hub.NewClient(h, conn)
		h.Register(client)

		// Send time sync
		client.Send(hub.TimeSyncMessage())

		go client.WritePump()
		go client.ReadPump()
	}
}
