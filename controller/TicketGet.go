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

	ticketIDStr := r.FormValue("ticket_id")
	if ticketIDStr == "" {
		http.Error(w, "ticket_id is required", http.StatusBadRequest)
		return
	}

	ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
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
	fmt.Printf(" ticket %s\n", ticket)
	var childObjectIDs []primitive.ObjectID
	var children [][]interface{}
	for _, childIDStr := range ticket.ChildrenIDs {
		childObjectID, err := primitive.ObjectIDFromHex(childIDStr)
		if err == nil {
			childObjectIDs = append(childObjectIDs, childObjectID)
		}
	}

	var childTickets []collection.Ticket

	if len(childObjectIDs) > 0 {
		filter = bson.M{"_id": bson.M{"$in": childObjectIDs}}
		cursor, err := coll.Find(ctx, filter)
		if err != nil {
			http.Error(w, "Failed to fetch child tickets", http.StatusInternalServerError)
			return
		}

		// Decode all the child tickets into the slice
		if err = cursor.All(ctx, &childTickets); err != nil {
			http.Error(w, "Failed to decode child tickets", http.StatusInternalServerError)
			return
		}

		// Build the children response with ID and Title if child tickets exist
		for _, childTicket := range childTickets {
			children = append(children, []interface{}{childTicket.TicketID.Hex(), childTicket.Title})
		}
	}

	response := struct {
		Ticket   collection.Ticket `json:"ticket"`
		Children [][]interface{}   `json:"children"`
	}{
		Ticket:   ticket,
		Children: children,
	}

	// Respond with the ticket and children in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)


}