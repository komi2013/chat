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

  // webpush "github.com/SherClockHolmes/webpush-go"

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
	coll := db1.Collection("book_pattern")
	var shiftStaffs []struct {
		BookPatternID string   `json:"bookPatternID,omitempty"`
		AliasName     string   `json:"aliasName,omitempty"`
		ShiftStart    string   `json:"shiftStart,omitempty"`
		ShiftEnd      string   `json:"shiftEnd,omitempty"`
		Skills        []string `json:"skills,omitempty"`
		Seq           int      `json:"seq,omitempty"`
		Delete        bool      `json:"delete,omitempty"`
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
	filter := bson.M{
		"_id": bookPatternID,
	}
	var bookPattern collection.BookPatternStruct
	err = coll.FindOne(ctx, filter).Decode(&bookPattern)
	if err != nil {
		http.Error(w, "No matching document found", http.StatusNotFound)
		return
	}

	// Prepare a map to track existing TimeSlots by date for quick lookup
	timeSlotMap := make(map[string]*collection.TimeSlot)
	for _, slot := range bookPattern.Times {
		timeSlotMap[slot.Date] = &slot
	}

	for date, slot := range timeSlotMap {
		log.Printf("スタートDate: %s, TimeSlot: %+v", date, *slot)
	}
	for _, staff := range shiftStaffs {
		date := staff.ShiftStart[:10]
		slot, exists := timeSlotMap[date]
		log.Printf("slot begin: %v", slot)
		if !exists {
			slot = &collection.TimeSlot{
				Date:        date,
				LimitStart:  "", // Set default values if needed
				LimitEnd:    "",
				ShiftStaff:  []collection.ShiftStaff{},
				Book:        []collection.Book{},
			}
		}
		existingMap := make(map[string]collection.ShiftStaff)
		for _, s := range slot.ShiftStaff {
			key := fmt.Sprintf("%s|%s", s.AliasName, s.ShiftStart)
			reqKey := fmt.Sprintf("%s|%s", staff.AliasName, staff.ShiftStart[11:])
			if key == reqKey && staff.Delete {
				log.Printf("Skipping staff %s on date %s due to delete flag", staff.AliasName, date)
				continue
			}
			existingMap[key] = s
		}
		key := fmt.Sprintf("%s|%s", staff.AliasName, staff.ShiftStart[11:])
		existingMap[key] = collection.ShiftStaff{
			AliasName:  staff.AliasName,
			ShiftStart: staff.ShiftStart[11:], // Extract only the HH:MM part
			ShiftEnd:   staff.ShiftEnd[11:],   // Extract only the HH:MM part
			Skills:     staff.Skills,
			Seq:        staff.Seq,
		}
		updatedShiftStaff := make([]collection.ShiftStaff, 0, len(existingMap))
		for _, s := range existingMap {
			updatedShiftStaff = append(updatedShiftStaff, s)
		}
		slot.ShiftStaff = updatedShiftStaff
		log.Printf("slot end: %v", slot)
	}
	for date, slot := range timeSlotMap {
		log.Printf("エンドDate: %s, TimeSlot: %+v", date, *slot)
	}
	// Rebuild bookPattern.Times from the updated map
	updatedTimes := make([]collection.TimeSlot, 0, len(timeSlotMap))
	for _, slot := range timeSlotMap {
		updatedTimes = append(updatedTimes, *slot)
	}

	// Assign updatedTimes back to bookPattern
	bookPattern.Times = updatedTimes

	update := bson.M{
		"$set": bson.M{
			"times": updatedTimes,
		},
	}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update shift staff", http.StatusInternalServerError)
		log.Printf("Update error: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Shift staffs updated successfully")
}

func toStringSlice(input interface{}) []string {
	if input == nil {
		return nil
	}
	if items, ok := input.([]interface{}); ok {
		result := make([]string, len(items))
		for i, v := range items {
			result[i] = v.(string)
		}
		return result
	}
	return nil
}

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

// func ShiftStaffEdit(w http.ResponseWriter, r *http.Request) {
// 	session, err := common.Session(w,r)
// 	if err != nil {
//   	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
//     return
// 	}
//   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//   defer cancel()
//   c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
//   if err != nil {
//     log.Print(err)
//   }
//   defer c.Disconnect(ctx)
//   db1 := c.Database(common.MongoDb1)

//   verifiedName := ""
//   for _, aliasArray := range session.AliasArray {
//   	if (r.FormValue("aliasName") == aliasArray[0]) {
//   		verifiedName = r.FormValue("aliasName")
//   	}
//   }
//   if verifiedName == "" {
//   	fmt.Printf("verifiedName err %s\n", r.FormValue("aliasName"))
//   	return
//   }
// 	var shiftStaffs []collection.ShiftStaffStruct
// 	err = json.Unmarshal([]byte(r.FormValue("contents")), &shiftStaffs)
// 	if err != nil {
// 		fmt.Printf("shiftStaffs err %s\n", r.FormValue("contents"))
// 		return
// 	}
// 	// [{"shiftStaffID":"ncsW202411011000sei1","bookPatternID":"ncsW","aliasName":"sei1","shiftStart":"2024-11-01T10:00","shiftEnd":"2024-11-01T15:00","role":"stylist","seq":1}]
// 	coll := db1.Collection("shift_staff")
// 	for _, shiftStaff := range shiftStaffs {
// 		filter := bson.M{"_id": shiftStaff.ShiftStaffID}
// 		if shiftStaff.Delete == 1 {
// 			_, err := coll.DeleteOne(context.TODO(), filter)
// 			if err != nil {
// 				log.Printf("Failed to delete shiftStaff with ID %s: %v", shiftStaff.ShiftStaffID, err)
// 			}
// 		} else {
// 			upsert := bson.M{
// 				"$set": bson.M{
// 					"channel_id":   r.FormValue("channelID"),
// 					"book_pattern_id":   shiftStaff.BookPatternID,
// 					"alias_name":  shiftStaff.AliasName,
// 					"shift_start": shiftStaff.ShiftStart,
// 					"shift_end":   shiftStaff.ShiftEnd,
// 					"role":        shiftStaff.Role,
// 					"seq":         shiftStaff.Seq,
// 				},
// 			}

// 			opts := options.Update().SetUpsert(true)
// 			_, err := coll.UpdateOne(context.TODO(), filter, upsert, opts)
// 			if err != nil {
// 				log.Printf("Failed to upsert shiftStaff with ID %s: %v", shiftStaff.ShiftStaffID, err)
// 			}
// 		}
// 	}

//   fmt.Fprint(w, `{"Status":"1"}`)
// }
