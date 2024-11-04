package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ShiftStaffGet(w http.ResponseWriter, r *http.Request) {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  coll := db1.Collection("shift_staff")
  filter := bson.M{
  	"channel_id": r.FormValue("channelID"),
  	"book_pattern_id": r.FormValue("bookPatternID"),
  }
  project := bson.D{
  	{"_id", 1},
  	{"alias_name", 1},
  	{"shift_start", 1},
  	{"shift_end", 1},
  	{"role", 1},
  	{"seq", 1},
  	{"channelID", -1},
  	{"bookPatternID", -1},
  	{"delete", -1},
  }
  opts := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts)
  if err != nil {
    fmt.Printf(" shift_staff find err %s\n", err)
  }
  var shiftStaffs []collection.ShiftStaffStruct
  if err = cursor.All(context.TODO(), &shiftStaffs); err != nil {
    fmt.Printf(" shiftStaffs all err %s\n", err)
    return
  }

	if len(shiftStaffs) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

  w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(shiftStaffs); err != nil {
		http.Error(w, "Failed to encode shiftStaffs to JSON", http.StatusInternalServerError)
	}

  // fmt.Fprint(w, shiftStaffs)
}

