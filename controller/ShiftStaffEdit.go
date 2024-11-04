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
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

// Define the structs for the parent structure
// type TimeEntry struct {
// 	BookTitle   string            `json:"bookTitle"`
// 	Date        string            `json:"date"`
// 	LimitStart  string            `json:"limitStart"`
// 	LimitEnd    string            `json:"limitEnd"`
// 	NeedRoles   [][2]interface{}  `json:"needRoles"` // Assuming [int, string] in a mixed slice
// }

// type Parent struct {
// 	AdminGroup    string          `json:"adminGroup"`
// 	NeedFacilities [][2]interface{} `json:"needFacilities"`
// 	MaxFacility   string          `json:"maxFacility"`
// 	Times         []TimeEntry     `json:"times"`
// 	BookPatternID string          `json:"bookPatternID"`
// 	OpenTimes     [][3]string     `json:"openTimes,omitempty"`
// }

func ShiftStaffEdit(w http.ResponseWriter, r *http.Request) {
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

  verifiedName := ""
  for _, aliasArray := range session.AliasArray {
  	if (r.FormValue("aliasName") == aliasArray[0]) {
  		verifiedName = r.FormValue("aliasName")
  	}
  }
  if verifiedName == "" {
  	fmt.Printf("verifiedName err %s\n", r.FormValue("aliasName"))
  	return
  }
	var shiftStaffs []collection.ShiftStaffStruct
	err = json.Unmarshal([]byte(r.FormValue("contents")), &shiftStaffs)
	if err != nil {
		fmt.Printf("shiftStaffs err %s\n", r.FormValue("contents"))
		return
	}
// [{"shiftStaffID":"ncsW202411011000sei1","bookPatternID":"ncsW","aliasName":"sei1","shiftStart":"2024-11-01T10:00","shiftEnd":"2024-11-01T15:00","role":"stylist","seq":1}]
	coll := db1.Collection("shift_staff")
	for _, shiftStaff := range shiftStaffs {
		filter := bson.M{"_id": shiftStaff.ShiftStaffID}
		if shiftStaff.Delete == 1 {
			_, err := coll.DeleteOne(context.TODO(), filter)
			if err != nil {
				log.Printf("Failed to delete shiftStaff with ID %s: %v", shiftStaff.ShiftStaffID, err)
			}
		} else {
			upsert := bson.M{
				"$set": bson.M{
					"channel_id":   r.FormValue("channelID"),
					"book_pattern_id":   shiftStaff.BookPatternID,
					"alias_name":  shiftStaff.AliasName,
					"shift_start": shiftStaff.ShiftStart,
					"shift_end":   shiftStaff.ShiftEnd,
					"role":        shiftStaff.Role,
					"seq":         shiftStaff.Seq,
				},
			}

			opts := options.Update().SetUpsert(true)
			_, err := coll.UpdateOne(context.TODO(), filter, upsert, opts)
			if err != nil {
				log.Printf("Failed to upsert shiftStaff with ID %s: %v", shiftStaff.ShiftStaffID, err)
			}
		}
	}

  fmt.Fprint(w, `{"Status":"1"}`)
}
