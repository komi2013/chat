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
  "go.mongodb.org/mongo-driver/bson/primitive"

  // "github.com/mitchellh/mapstructure"

  "chat/collection"
  "chat/common"
)

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

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  trueAccess := false
  for _, arrayData := range session.AliasArray {
    if arrayData[0] == aliasName && arrayData[1] == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    fmt.Printf(" err %s\n", session.AliasArray, aliasName)
    return
  }

	var shiftStaffs []struct {
		BookPatternID string   `json:"bookPatternID,omitempty"`
		AliasName     string   `json:"aliasName,omitempty"`
		ShiftStart    string   `json:"shiftStart,omitempty"`
		ShiftEnd      string   `json:"shiftEnd,omitempty"`
		Skills        []string `json:"skills,omitempty"`
		Seq           int      `json:"seq,omitempty"`
		Delete        bool     `json:"delete,omitempty"`
		ShiftStaffID  string   `json:"shiftStaffID,omitempty"`
	}
	if err := json.Unmarshal([]byte(r.FormValue("shiftStaffs")), &shiftStaffs); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		log.Printf("JSON error: %v", err)
		return
	}
	if len(shiftStaffs) == 0 {
		http.Error(w, "No shiftStaffs provided", http.StatusBadRequest)
		return
	}
	firstStaff := shiftStaffs[0]
	bookPatternID, err := primitive.ObjectIDFromHex(firstStaff.BookPatternID)
	if err != nil {
		http.Error(w, "Invalid bookPatternID format", http.StatusBadRequest)
		return
	}
	coll := db1.Collection("book_pattern")
	var bookPattern collection.BookPatternStruct
	filter := bson.M{"_id": bookPatternID}
	for attempt := 0; attempt < 3; attempt++ {
		update := bson.M{
			"$set": bson.M{"lock": true},
		}
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err := coll.FindOneAndUpdate(ctx, bson.M{
			"$and": []bson.M{
				filter,
				{"lock": bson.M{"$ne": true}},
			},
		}, update, options).Decode(&bookPattern)
		if err == nil {
			break
		} else if err == mongo.ErrNoDocuments {
			log.Printf("Attempt %d: Document is locked, retrying...", attempt+1)
			time.Sleep(1 * time.Second)
			continue
		} else {
			log.Printf("Failed to lock and retrieve document: %v", err)
			return
		}
		if attempt == 2 {
			log.Println("Failed to acquire lock after 3 attempts")
			return
		}
	}
	timeSlotMap := make(map[string]*collection.TimeSlot)
	for i := range bookPattern.Times {
		slot := &bookPattern.Times[i]
		timeSlotMap[slot.Date] = slot
	}

	for _, staff := range shiftStaffs {
		date := staff.ShiftStart[:10]
		slot, exists := timeSlotMap[date]
		if !exists {
			newSlot := &collection.TimeSlot{
				Date:        date,
				LimitStart:  "",
				LimitEnd:    "",
				ShiftStaffs:  []collection.ShiftStaffs{},
				Books:        []collection.Books{},
			}
			timeSlotMap[date] = newSlot
			slot = newSlot
		}
		updatedShiftStaff := []collection.ShiftStaffs{}
		for _, s := range slot.ShiftStaffs {
			key := fmt.Sprintf("%s|%s", s.AliasName, s.ShiftStart)
			reqKey := fmt.Sprintf("%s|%s", staff.AliasName, staff.ShiftStart[11:])
			if key == reqKey && staff.Delete {
				continue
			}
			updatedShiftStaff = append(updatedShiftStaff, s)
		}
		if !staff.Delete {
			updatedShiftStaff = append(updatedShiftStaff, collection.ShiftStaffs{
				AliasName:  staff.AliasName,
				ShiftStart: staff.ShiftStart[11:],
				ShiftEnd:   staff.ShiftEnd[11:],
				Skills:     staff.Skills,
				Seq:        staff.Seq,
			})
		}
		slot.ShiftStaffs = updatedShiftStaff
	}
	updatedTimes := make([]collection.TimeSlot, 0, len(timeSlotMap))
	for _, slot := range timeSlotMap {
		updatedTimes = append(updatedTimes, *slot)
	}
	bookPattern.Times = updatedTimes
	update := bson.M{
		"$set": bson.M{
			"times": updatedTimes,
		},
		"$unset": bson.M{"lock": ""},
	}
	opts := options.Update().SetUpsert(false)
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Failed to update shift staff", http.StatusInternalServerError)
		log.Printf("Update error: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Shift staffs updated successfully")
}

