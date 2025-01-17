package common

import (
  "encoding/json"
  "log"
  "net/http"
  "unicode/utf8"

  "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
)

func ChunkPush(
  sessions []collection.SessionStruct,
  db1 *mongo.Database,
  r *http.Request,
  arr []interface{},
) {
	jsonData, _ := json.Marshal(arr)
  const maxChunkSize = 2000 / utf8.UTFMax

  chunk := false
  chunks := []string{string(jsonData)}
  if len(jsonData) > 2000 {
    chunk = true
    chunks = SplitIntoByteChunks(string(jsonData), maxChunkSize)
  }

  chunkLength := len(chunks)
  chunkPass := StringRand(2)

  for chIndex, ch := range chunks {
    for _, session := range sessions {
      pushID := StringRand(12)
      var arrForJson []interface{}

      if chunk {
        arrForJson = []interface{}{
          pushID, "chunk", ch, chunkPass, chIndex, chunkLength,
        }
      } else {
        arrForJson = append([]interface{}{pushID}, arr...)
      }
      resp, err := SendWebPushNotification(db1, arrForJson, pushID, session.Subscription)
      if err != nil {
        log.Printf("Error in SendWebPushNotification: %v; Req: %s", err, r.URL.Path)
      } else {
        defer resp.Body.Close()
      }
    }
  }
}

