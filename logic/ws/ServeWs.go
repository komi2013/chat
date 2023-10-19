package ws

import (
  "log"
  "net/http"
)

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
  upgrader.CheckOrigin = func(r *http.Request) bool {
    return r.Header.Get("Origin") == "https://local.quigen.info"
  }
  // fmt.Printf("r.Header.Get('Origin') %s\n", r.Header.Get("Origin"))
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
    return
  }
  client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}

// after authorization, take user_id from session collection
// select alias_name from alias where user_id = ?
// upsert session 
// client.id = user_id
// client.alias_names = [alias_name1, alias_name2]
// 


  cookie, _ := r.Cookie("ss")
  client.id = cookie.Value
  client.hub.register <- client

  // Allow collection of memory referenced by the caller by doing all work in
  // new goroutines.
  go client.writePump()
  go client.readPump()
}
