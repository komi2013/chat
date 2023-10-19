package ws

import (
  "encoding/json"
  "fmt"
)

func (h *Hub) Run() {
  for {
    select {
    case client := <-h.register:
      h.clients[client] = true
      // fmt.Printf(" r %s\n", r)
      fmt.Printf(" client register %s\n", client.id)

      m1 := []interface{}{"init"}
      s, _ := json.Marshal(m1)
      fmt.Printf(" s %s\n", string(s))

      client.send <- []byte(string(s))
    case client := <-h.unregister:
      if _, ok := h.clients[client]; ok {
        delete(h.clients, client)
        close(client.send)
      }
    case message := <-h.broadcast:
      // check sender correct chanel id for secure
      // select user which has same chnnel id
      for client := range h.clients {
// if it is first time reload
// select * from channel where channel_id in (
//   select channel_id from channel_group where alias_name in client.asias_names
//    ) 
        var m1 []interface{}
        if err := json.Unmarshal(message, &m1); err != nil {
            // log.Fatal(err)
            fmt.Printf(" err %s\n", err)
        }
        fmt.Printf(" m1 %s\n", m1)
        // if client.id == "123" {
          fmt.Printf(" client.id %s\n", client.id)
          fmt.Printf(" h.broadcast %s\n", h.broadcast)
          select {
          // case client.send <- []byte("bytedance"):
          case client.send <- message:
          default:
            close(client.send)
            delete(h.clients, client)
          }
        // }
      }
    }
  }
}
