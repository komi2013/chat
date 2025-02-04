package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  // "github.com/mitchellh/mapstructure"

  "chat/collection"
  "chat/common"
)

func BookAdd(w http.ResponseWriter, r *http.Request) {
  _, err := common.Session(w,r)
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

	bookStart := r.FormValue("bookStart")
	bookEnd := r.FormValue("bookEnd")
	serviceID, err := strconv.Atoi(r.FormValue("serviceID"))
	if err != nil {
		http.Error(w, "Invalid serviceID format", http.StatusBadRequest)
		return
	}

  var answers []string
  if err := json.Unmarshal([]byte(r.FormValue("answers")), &answers); err != nil {
    http.Error(w, "Invalid JSON answers", http.StatusBadRequest)
    return
  }
  bookPatternID, err := primitive.ObjectIDFromHex(r.FormValue("bookPatternID"))
	if err != nil {
		http.Error(w, "Invalid bookPatternID format", http.StatusBadRequest)
		return
	}

	if bookStart == "" || bookEnd == "" {
		http.Error(w, "Missing bookStart or bookEnd", http.StatusBadRequest)
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

	layout := "2006-01-02T15:04"
	bookStartTime, err := time.Parse(layout, bookStart)
	if err != nil {
		return
	}
	bookEndTime, err := time.Parse(layout, bookEnd)
	if err != nil {
		return
	}

	var requestedService *collection.Service
	for _, service := range bookPattern.Services {
		if service.ID == serviceID {
			requestedService = &service
			break
		}
	}
	if requestedService == nil {
		log.Printf("Service ID %d not found", serviceID)
		return
	}

	updatedTimes := make([]collection.TimeSlot, len(bookPattern.Times))
	copy(updatedTimes, bookPattern.Times)
	bookingAdded := false

	for i, timeSlot := range updatedTimes {
		if timeSlot.Date != bookStartTime.Format("2006-01-02") {
			continue
		}

		staffAvailable := false
		for _, staff := range timeSlot.ShiftStaffs {
			staffStart, _ := time.Parse("15:04", staff.ShiftStart)
			staffEnd, _ := time.Parse("15:04", staff.ShiftEnd)

			shiftStart := time.Date(bookStartTime.Year(), bookStartTime.Month(), bookStartTime.Day(), staffStart.Hour(), staffStart.Minute(), 0, 0, bookStartTime.Location())
			shiftEnd := time.Date(bookEndTime.Year(), bookEndTime.Month(), bookEndTime.Day(), staffEnd.Hour(), staffEnd.Minute(), 0, 0, bookEndTime.Location())
			if !bookStartTime.Before(shiftStart) && bookEndTime.Before(shiftEnd) &&
			    (requestedService.NeedSkill == "" || contains(staff.Skills, requestedService.NeedSkill)) {
			    staffAvailable = true
			    break
			}
		}

		if staffAvailable {
			overlaps := false
			for _, booking := range timeSlot.Books {
				bookedStart, _ := time.Parse("15:04", booking.BookStart)
				bookedEnd, _ := time.Parse("15:04", booking.BookEnd)

				existingStart := time.Date(bookStartTime.Year(), bookStartTime.Month(), bookStartTime.Day(), bookedStart.Hour(), bookedStart.Minute(), 0, 0, bookStartTime.Location())
				existingEnd := time.Date(bookEndTime.Year(), bookEndTime.Month(), bookEndTime.Day(), bookedEnd.Hour(), bookedEnd.Minute(), 0, 0, bookEndTime.Location())
				if (bookEndTime.After(existingStart) && bookStartTime.Before(existingEnd)) {
					overlaps = true
					break
				}
			}

			if !overlaps {
				newBooking := collection.Books{
					BookStart: bookStartTime.Format("15:04"),
					BookEnd:   bookEndTime.Format("15:04"),
					Answers:   answers,
					ServiceID: serviceID,
				}
				updatedTimes[i].Books = append(updatedTimes[i].Books, newBooking)
				bookingAdded = true
				break
			}
		}
	}

	if bookingAdded {
		if requestedService.PrepaidPrice == 0 {
			bookingAdded = false
			// if requestedService.PrepaidPrice > ss.Yen {
			// 	bookingAdded = false
			// }
		}
	} else {
		log.Printf("bookingAdded: %v", bookingAdded)
	}

	update := bson.M{
		"$unset": bson.M{"lock": ""},
	}

	if bookingAdded {
		update["$set"] = bson.M{
			"times": updatedTimes,
		}
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

// func AddBooking(ctx context.Context, db *mongo.Database, collectionName string, bookPatternID primitive.ObjectID, bookStart string, bookEnd string, menuID int, answers []string) (bool, error) {
// 	collection := db.Collection(collectionName)

// 	// Find the existing BookPatternStruct
// 	var bookPattern collection.BookPatternStruct
// 	err := collection.FindOne(ctx, bson.M{"_id": bookPatternID}).Decode(&bookPattern)
// 	if err != nil {
// 		return false, err
// 	}

// 	layout := "2006-01-02T15:04"
// 	bookStartTime, err := time.Parse(layout, bookStart)
// 	if err != nil {
// 		return false, err
// 	}
// 	bookEndTime, err := time.Parse(layout, bookEnd)
// 	if err != nil {
// 		return false, err
// 	}

// 	// Find the requested menu
// 	var requestedMenu *collection.MenuItem
// 	for _, menu := range bookPattern.Menus {
// 		if menu.ID == menuID {
// 			requestedMenu = &menu
// 			break
// 		}
// 	}
// 	if requestedMenu == nil {
// 		log.Printf("Menu ID %d not found", menuID)
// 		return false, nil
// 	}

// 	// Initialize updatedTimes
// 	updatedTimes := make([]collection.TimeSlot, len(bookPattern.Times))
// 	copy(updatedTimes, bookPattern.Times)

// 	bookingAdded := false

// 	// Process each time slot
// 	for i, timeSlot := range updatedTimes {
// 		if timeSlot.Date != bookStartTime.Format("2006-01-02") {
// 			continue
// 		}

// 		staffAvailable := false
// 		for _, staff := range timeSlot.ShiftStaffs {
// 			staffStart, _ := time.Parse("15:04", staff.ShiftStart)
// 			staffEnd, _ := time.Parse("15:04", staff.ShiftEnd)

// 			shiftStart := time.Date(bookStartTime.Year(), bookStartTime.Month(), bookStartTime.Day(), staffStart.Hour(), staffStart.Minute(), 0, 0, bookStartTime.Location())
// 			shiftEnd := time.Date(bookEndTime.Year(), bookEndTime.Month(), bookEndTime.Day(), staffEnd.Hour(), staffEnd.Minute(), 0, 0, bookEndTime.Location())

// 			// Check skill and time availability
// 			if bookStartTime.After(shiftStart) && bookEndTime.Before(shiftEnd) &&
// 				(requestedMenu.NeedSkill == "" || contains(staff.Skills, requestedMenu.NeedSkill)) {
// 				staffAvailable = true
// 				break
// 			}
// 		}

// 		// Check booking overlaps
// 		if staffAvailable {
// 			overlaps := false
// 			for _, booking := range timeSlot.Books {
// 				bookedStart, _ := time.Parse("15:04", booking.BookStart)
// 				bookedEnd, _ := time.Parse("15:04", booking.BookEnd)

// 				existingStart := time.Date(bookStartTime.Year(), bookStartTime.Month(), bookStartTime.Day(), bookedStart.Hour(), bookedStart.Minute(), 0, 0, bookStartTime.Location())
// 				existingEnd := time.Date(bookEndTime.Year(), bookEndTime.Month(), bookEndTime.Day(), bookedEnd.Hour(), bookedEnd.Minute(), 0, 0, bookEndTime.Location())

// 				if !(bookEndTime.Before(existingStart) || bookStartTime.After(existingEnd)) {
// 					overlaps = true
// 					break
// 				}
// 			}

// 			// Add booking if no overlap
// 			if !overlaps {
// 				newBooking := collection.Books{
// 					BookStart: bookStartTime.Format("15:04"),
// 					BookEnd:   bookEndTime.Format("15:04"),
// 					Answers:   answers,
// 					MenuID:    menuID,
// 				}
// 				updatedTimes[i].Books = append(updatedTimes[i].Books, newBooking)
// 				bookingAdded = true
// 				break
// 			}
// 		}
// 	}

// 	if !bookingAdded {
// 		log.Printf("bookingAdded: %v", bookingAdded)
// 		return false, nil
// 	}

// 	// Update the MongoDB document
// 	update := bson.M{
// 		"$set": bson.M{
// 			"times": updatedTimes,
// 		},
// 		"$unset": bson.M{"lock": ""},
// 	}
// 	opts := options.Update().SetUpsert(false)

// 	_, err = collection.UpdateOne(ctx, bson.M{"_id": bookPatternID}, update, opts)
// 	if err != nil {
// 		return false, err
// 	}

// 	log.Printf("Booking successfully added")
// 	return true, nil
// }

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}




func isTimeInRange(targetStart, targetEnd, rangeStart, rangeEnd, date string) bool {
	if date != targetStart[:10] {
		return false
	}
	layout := "15:04"
	tStart, _ := time.Parse(layout, targetStart)
	tEnd, _ := time.Parse(layout, targetEnd)
	rStart, _ := time.Parse(layout, rangeStart)
	rEnd, _ := time.Parse(layout, rangeEnd)

	return !tStart.Before(rStart) && !tEnd.After(rEnd)
}

func countBookings(bookings []collection.Books, targetStart, targetEnd string) int {
	layout := "15:04"
	tStart, _ := time.Parse(layout, targetStart)
	tEnd, _ := time.Parse(layout, targetEnd)

	count := 0
	for _, booking := range bookings {
		bStart, _ := time.Parse(layout, booking.BookStart)
		bEnd, _ := time.Parse(layout, booking.BookEnd)
		if !(tEnd.Before(bStart) || tStart.After(bEnd)) {
			count++
		}
	}
	return count
}
