package ws

import (
	"github.com/gorilla/websocket"

	// "chat/logic/hub"
)

type Client struct {
  // unique ID for each client
  id string

  hub *Hub

  // The websocket connection.
  conn *websocket.Conn

  // Buffered channel of outbound messages.
  send chan []byte
}