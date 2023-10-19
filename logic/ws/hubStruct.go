package ws

// import (
// 	"chat/logic/client"
// )

type Hub struct {
  // Registered clients.
  clients map[*Client]bool

  // Inbound messages from the clients.
  broadcast chan []byte

  // Register requests from the clients.
  register chan *Client

  // Unregister requests from clients.
  unregister chan *Client
}