package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func BookPatternShift(w http.ResponseWriter, r *http.Request) {

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  bookPatternID, err := primitive.ObjectIDFromHex(r.FormValue("bookPatternID"))
  if err != nil {
    http.Error(w, "Invalid bookPatternID format", http.StatusBadRequest)
    return
  }

	var skills []string
	if err := json.Unmarshal([]byte(r.FormValue("availableSkills")), &skills); err != nil {
		log.Printf("JSON Unmarshal Error: %v; Request: %v", err, r.Form)
		http.Error(w, "availableSkills Invalid JSON format", http.StatusBadRequest)
		return
	}

	type updatedShiftStruct struct {
		AliasNames   []string `bson:"alias_names" json:"aliasNames"`
		Date         string   `bson:"date" json:"date"`
		Start        string   `bson:"start" json:"start"`
		End          string   `bson:"end" json:"end"`
		Role         string   `bson:"role" json:"role"`
		AddMinusFlag int      `bson:"add_minus_flag" json:"addMinusFlag"`
		Confirmed    int      `bson:"confirmed" json:"confirmed"`
	}

	var updatedShifts []updatedShiftStruct
	if err := json.Unmarshal([]byte(r.FormValue("updatedShifts")), &updatedShifts); err != nil {
		log.Printf("JSON Unmarshal Error: %v; Request: %v", err, r.Form)
		http.Error(w, "updatedShifts Invalid JSON format", http.StatusBadRequest)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  trueAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      trueAccess = true
    }
  }
  if !trueAccess {
    log.Printf("ChannelAliases !trueAccess: %v; Req: ", session.ChannelAliases, r.URL.Path, r.Form)
    return
  }

	// shiftDate := "2025-03-08"             // shiftsAliasNames から取得
	// aliasName := "ivan1"                        // shiftsAliasNames から取得
	// newSkill := "cut"      
  coll := db1.Collection("book_pattern")

  // Find the document by `_id`
  var bookPattern collection.BookPatternStruct
  filter := bson.M{"_id": bookPatternID}
  err = coll.FindOne(ctx, filter).Decode(&bookPattern)
  if err != nil {
    log.Print(err, " bookPattern ", bookPatternID)
    http.Error(w, "Book pattern not found", http.StatusNotFound)
    return
  }
  log.Print(" bookPattern ", bookPatternID)

	// **2. updatedShifts のデータをループ処理**
	for _, shift := range updatedShifts {
		shiftDate := shift.Date

		// **3. TimeSlot の中で shiftDate に一致するものを探す**
		var timeSlot *collection.TimeSlot
		for i := range bookPattern.Times {
			if bookPattern.Times[i].Date == shiftDate {
				timeSlot = &bookPattern.Times[i]
				break
			}
		}

		// **4. TimeSlot が存在しない場合は新規追加**
		if timeSlot == nil {
			log.Printf("TimeSlot not found for date: %s, creating new one...", shiftDate)

			newTimeSlot := collection.TimeSlot{
				Date:       shiftDate,
				LimitStart: shift.Start, // 初期値として start を設定
				LimitEnd:   shift.End,   // 初期値として end を設定
				WorkStaffs: []collection.WorkStaff{}, // 空のスタッフリスト
			}

			// `$push` で `times` 配列に新しい `TimeSlot` を追加
			pushUpdate := bson.M{
				"$push": bson.M{
					"times": newTimeSlot,
				},
			}

			_, err := coll.UpdateOne(ctx, filter, pushUpdate)
			if err != nil {
				log.Print(err, " Failed to insert new TimeSlot")
				continue
			}

			// 追加後、bookPattern を再取得して新しい TimeSlot を設定
			err = coll.FindOne(ctx, filter).Decode(&bookPattern)
			if err != nil {
				log.Print(err, " Failed to reload BookPattern")
				continue
			}

			// もう一度、TimeSlot を探す
			for i := range bookPattern.Times {
				if bookPattern.Times[i].Date == shiftDate {
					timeSlot = &bookPattern.Times[i]
					break
				}
			}
		}

		// **4. aliasNames をループして更新**
		for _, aliasName := range shift.AliasNames {
			update := bson.M{
				"$set": bson.M{
					"times.$[].work_staffs.$[staff].skills": skills, // スキルを更新
				},
				"$setOnInsert": bson.M{ // 存在しなければ新規追加
					"times.$[].work_staffs.$[staff].alias_name": aliasName,
					"times.$[].work_staffs.$[staff].work_start": shift.Start,
					"times.$[].work_staffs.$[staff].work_end":   shift.End,
					"times.$[].work_staffs.$[staff].seq":        len(timeSlot.WorkStaffs) + 1,
				},
			}

			arrayFilters := options.ArrayFilters{
				Filters: []interface{}{
					bson.M{"staff.alias_name": aliasName},
				},
			}

			updateOptions := options.Update().SetArrayFilters(arrayFilters).SetUpsert(true)

			result, err := coll.UpdateOne(ctx, filter, update, updateOptions)
			if err != nil {
				log.Print(err, " WorkStaff update/insert error")
				continue
			}
			log.Printf("Updated or Inserted %d document(s) for alias: %s", result.ModifiedCount, aliasName)
		}

		log.Print(" before update shifts")

		shiftFilter := bson.M{
			"_id": bookPatternID,
			"shifts": bson.M{
				"$elemMatch": bson.M{
					"date":        shift.Date,
					"shift_start": shift.Start,
					"role":        shift.Role,
				},
			},
		}

		log.Printf("shiftFilter: %v", shiftFilter)

		shiftUpdate := bson.M{
			"$addToSet": bson.M{
				"shifts.$.alias_names": bson.M{"$each": shift.AliasNames}, 
			},
		}

		shiftResult, err := coll.UpdateOne(ctx, shiftFilter, shiftUpdate)
		if err != nil {
			log.Print(err, " Shifts alias_names update error")
			continue
		}
		log.Printf("Updated %d shifts for date: %s, start: %s, role: %s", shiftResult.ModifiedCount, shift.Date, shift.Start, shift.Role)

	}

	// fmt.Fprintf(w, "WorkStaff updaskillstes completed")

  responseData := struct {
    Csrf         string        `json:"csrf"`
    PushContents []string `json:"pushContents"`
    BookPattern collection.BookPatternStruct `json:"bookPattern"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    BookPattern: bookPattern,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)


}
