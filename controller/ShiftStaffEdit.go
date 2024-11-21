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

	coll := db1.Collection("book_pattern")
	var bookPattern collection.BookPatternStruct
	filter := bson.M{"_id": bookPatternID}

	// Retry logic to attempt acquiring the lock
	for attempt := 0; attempt < 3; attempt++ {
		// Atomic update to check and acquire lock
		update := bson.M{
			"$set": bson.M{"lock": true}, // Set the lock
		}
		options := options.FindOneAndUpdate().SetReturnDocument(options.After)

		// Try to find and lock the document
		err := coll.FindOneAndUpdate(ctx, bson.M{
			"$and": []bson.M{
				filter,
				{"lock": bson.M{"$ne": true}}, // Ensure the document is not already locked
			},
		}, update, options).Decode(&bookPattern)

		if err == nil {
			// Successfully acquired the lock and decoded the document
			break
		} else if err == mongo.ErrNoDocuments {
			// Document is already locked; retry after 1 second
			log.Printf("Attempt %d: Document is locked, retrying...", attempt+1)
			time.Sleep(1 * time.Second)
			continue
		} else {
			// Other errors, log and return
			log.Printf("Failed to lock and retrieve document: %v", err)
			return
		}

		// If retries are exhausted, log and return
		if attempt == 2 {
			log.Println("Failed to acquire lock after 3 attempts")
			return
		}
	}

	// Successfully locked and retrieved the document
	log.Printf("Locked and retrieved bookPattern: %+v", bookPattern)


	timeSlotMap := make(map[string]*collection.TimeSlot)
	for i := range bookPattern.Times {
		slot := &bookPattern.Times[i] // Reference the actual TimeSlot in bookPattern.Times
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
				ShiftStaff:  []collection.ShiftStaff{},
				Book:        []collection.Book{},
			}
			timeSlotMap[date] = newSlot
			slot = newSlot
		}
		updatedShiftStaff := []collection.ShiftStaff{}
		for _, s := range slot.ShiftStaff {
			key := fmt.Sprintf("%s|%s", s.AliasName, s.ShiftStart)
			reqKey := fmt.Sprintf("%s|%s", staff.AliasName, staff.ShiftStart[11:])
			if key == reqKey && staff.Delete {
				continue
			}
			updatedShiftStaff = append(updatedShiftStaff, s)
		}
		if !staff.Delete {
			updatedShiftStaff = append(updatedShiftStaff, collection.ShiftStaff{
				AliasName:  staff.AliasName,
				ShiftStart: staff.ShiftStart[11:], // Extract only the HH:MM part
				ShiftEnd:   staff.ShiftEnd[11:],   // Extract only the HH:MM part
				Skills:     staff.Skills,
				Seq:        staff.Seq,
			})
		}
		slot.ShiftStaff = updatedShiftStaff
		log.Printf("Updated slot: %+v", *slot)
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

	// filter := bson.D{{"_id", cookie.Value}}

	// lockUpdate := bson.M{"$unset": bson.M{"lock": ""}}

	// update := bson.D{{"$set", bson.D{
	// 	{"subscription", r.FormValue("subscription")},
	// 	{"updated_at", time.Now()}}}}
	opts := options.Update().SetUpsert(false)
	// _, err = coll.UpdateOne(context.TODO(), filter, update, opts)
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Failed to update shift staff", http.StatusInternalServerError)
		log.Printf("Update error: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Shift staffs updated successfully")
}

// CheckAndLock attempts to acquire a lock on a document.
// It retries for up to 3 seconds if the document is already locked.
func CheckAndLock(coll *mongo.Collection, filter bson.M, ctx context.Context) (bson.M, error) {
	// Define the update to acquire the lock
	update := bson.M{
		"$set": bson.M{"lock": true}, // Set the lock
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result bson.M

	// Retry logic: Try up to 3 times to acquire the lock
	for attempt := 0; attempt < 3; attempt++ {
		// Perform an atomic FindOneAndUpdate operation
		err := coll.FindOneAndUpdate(ctx, bson.M{
			"$and": []bson.M{
				filter,
				{"lock": bson.M{"$ne": true}}, // Ensure the document is not already locked
			},
		}, update, options).Decode(&result)

		if err == nil {
			// Successfully locked the document
			return result, nil
		} else if err == mongo.ErrNoDocuments {
			// Document is locked, retry after 1 second
			log.Printf("Attempt %d: Document is locked, retrying...", attempt+1)
			time.Sleep(1 * time.Second)
			continue
		} else {
			// Other errors, return immediately
			log.Printf("Failed to acquire lock: %v", err)
			return nil, fmt.Errorf("failed to lock document: %w", err)
		}
	}

	// If retries are exhausted, return an error
	log.Println("Failed to acquire lock after 3 attempts")
	return nil, fmt.Errorf("document is locked after 3 attempts")
}

// func ProcessDocument(coll *mongo.Collection, bookPatternID string, ctx context.Context) error {
// 	// Filter to identify the document by ID
// 	filter := bson.M{"_id": bookPatternID}

// 	// Atomic update to check for lock and acquire it if not already locked
// 	update := bson.M{
// 		"$set": bson.M{"lock": true}, // Set the lock
// 	}
// 	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

// 	// Attempt to acquire the lock
// 	var bookPattern collection.BookPatternStruct
// 	err := coll.FindOneAndUpdate(ctx, bson.M{
// 		"$and": []bson.M{
// 			filter,
// 			{"lock": bson.M{"$ne": true}}, // Ensure the document is not already locked
// 		},
// 	}, update, options).Decode(&bookPattern)

// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			// Document is already locked
// 			log.Println("Document is already locked by another process")
// 			return fmt.Errorf("document is locked")
// 		}
// 		// Other errors
// 		log.Fatalf("Failed to lock document: %v", err)
// 		return fmt.Errorf("failed to lock document: %w", err)
// 	}

// 	// Log that the document is locked and being processed
// 	log.Printf("Processing document: %+v", bookPattern)

// 	// Perform your processing logic here
// 	time.Sleep(2 * time.Second) // Simulated processing

// 	// Unlock the document after processing
// 	_, err = coll.UpdateOne(ctx, filter, bson.M{"$unset": bson.M{"lock": ""}})
// 	if err != nil {
// 		log.Printf("Failed to unlock document: %v", err)
// 		return fmt.Errorf("failed to unlock document: %w", err)
// 	}

// 	// Log that the document was successfully processed and unlocked
// 	log.Println("Document processed and unlocked successfully")
// 	return nil
// }


// Unlock removes the lock from the document.
// func Unlock(coll *mongo.Collection, filter bson.M, ctx context.Context) error {
// 	lockUpdate := bson.M{"$unset": bson.M{"lock": ""}}
// 	_, err := coll.UpdateOne(ctx, filter, lockUpdate)
// 	if err != nil {
// 		log.Printf("Failed to remove lock: %v", err)
// 		return fmt.Errorf("failed to unlock document: %w", err)
// 	}
// 	return nil
// }

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
