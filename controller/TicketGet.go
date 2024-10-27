package controller

import (
  "context"
  // "encoding/base64"
  "encoding/json"
  "fmt"
  // "io/ioutil"
  "log"
  "net/http"
  // "os"
  // "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func TicketGet(w http.ResponseWriter, r *http.Request) {
  session, err := common.Session(w,r)
  if err != nil {
    http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
  }
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName && arrayData[1] == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf("!trueAccess: %s\n", session.AliasArray, aliasName)
    return
  }

	ticketID, err := primitive.ObjectIDFromHex(r.FormValue("ticket_id"))
	if err != nil {
		http.Error(w, "Invalid ticket_id format", http.StatusBadRequest)
		return
	}

	coll := db1.Collection("ticket")
	var ticket collection.Ticket
	filter := bson.M{"_id": ticketID}
	err = coll.FindOne(ctx, filter).Decode(&ticket)
	if err != nil {
		log.Print(err, "ticket", r.FormValue("ticket_id"))
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}
	coll = db1.Collection("ticket_log")
	var ticketLogs []collection.TicketLog
	filter = bson.M{"ticket_id": ticketID}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
	    log.Printf("Error querying MongoDB: %v", err)
	    http.Error(w, "Error querying ticket logs", http.StatusInternalServerError)
	    return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &ticketLogs); err != nil {
	    log.Printf("Error decoding ticket logs: %v", err)
	    http.Error(w, "Error decoding ticket logs", http.StatusInternalServerError)
	    return
	}

	response := struct {
	    Ticket     collection.Ticket      `json:"ticket"`
	    TicketLogs []collection.TicketLog `json:"ticketLogs"`
	}{
	    Ticket:    ticket,
	    TicketLogs: ticketLogs, // this could be an empty slice if no logs were found
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}