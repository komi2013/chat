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
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionShift(w http.ResponseWriter, r *http.Request) {

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  bookPatternID := r.FormValue("bookPatternID")

	var skills []string
	if err := json.Unmarshal([]byte(r.FormValue("availableSkills")), &skills); err != nil {
		log.Printf("JSON Unmarshal Error: %v; Request: %v", err, r.Form)
		http.Error(w, "availableSkills Invalid JSON format", http.StatusBadRequest)
		return
	}

	var updatedShifts []collection.Shift
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

  coll := db1.Collection("book_pattern")

  // Find the document by `_id`
  var reception collection.ReceptionStruct
  filter := bson.M{"_id": bookPatternID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    log.Print(err, " reception ", bookPatternID)
    http.Error(w, "Book pattern not found", http.StatusNotFound)
    return
  }

	for _, shift := range updatedShifts {

    existingStaffMap := make(map[string]bool) // 既存の WorkStaff (AliasName → 存在判定)
    for _, staff := range reception.WorkStaffs {
      if staff.WorkStart == shift.ShiftStart {
        existingStaffMap[staff.AliasName] = true
      }
    }

    // `updatedShifts` の `AliasNames` から変化を取得
    newAliasMap := make(map[string]bool) // 新しい updatedShifts の AliasName (存在判定)
    for _, alias := range shift.AliasNames {
      newAliasMap[alias] = true
    }

    newOpenAliases := getOpenAliasesFromShift(shift.AliasNames, shift.Open)

    // ** 追加された AliasNames **
    var addedAliases []string
    for i, alias := range shift.AliasNames {
      if i < shift.Open && !existingStaffMap[alias] {
        addedAliases = append(addedAliases, alias)
      }
    }

    // ** 削除された AliasNames（updatedShifts にないもの + shift.Open から出たもの）**
    var removedAliases []string
    for alias := range existingStaffMap {
      if !newAliasMap[alias] || !common.SliceStrContains(newOpenAliases, alias) {
        removedAliases = append(removedAliases, alias)
      }
    }

    // ** 追加処理（shift.Open の範囲内のみ）**
    for _, aliasName := range addedAliases {
      // if addI < shift.Open {
      newStaff := collection.WorkStaff{
        AliasName: aliasName,
        WorkStart: shift.ShiftStart,
        WorkEnd:   shift.ShiftEnd,
        Seq:       len(reception.WorkStaffs) + 1,
      }
      reception.WorkStaffs = append(reception.WorkStaffs, newStaff)
      // }
    }

    // ** 削除処理（removedAliases のデータを WorkStaffs から削除）**
    for j := 0; j < len(reception.WorkStaffs); {
      staff := reception.WorkStaffs[j]
      if common.SliceStrContains(removedAliases, staff.AliasName) && staff.WorkStart == shift.ShiftStart {
        // 削除
        reception.WorkStaffs = append(reception.WorkStaffs[:j], reception.WorkStaffs[j+1:]...)
      } else {
        j++
      }
    }

		log.Printf("ShiftStart: %s", shift.ShiftStart)
		log.Printf("Added AliasNames: %v", addedAliases)
		log.Printf("Removed AliasNames: %v", removedAliases)

		for i, s := range reception.Shifts {
			if s.ShiftStart == shift.ShiftStart && s.Role == shift.Role {
				reception.Shifts[i].AliasNames = shift.AliasNames
				reception.Shifts[i].Fix = shift.Fix
			}
		}
	}

	log.Printf("Updated WorkStaffs: %+v", reception.WorkStaffs)

	if len(skills) > 0 {
		updated := false
		for j, staffSkill := range reception.StaffSkills {
			if staffSkill.AliasName == aliasName {
				reception.StaffSkills[j].Skills = skills
				updated = true
				break
			}
		}

		if !updated {
			newStaffSkill := collection.StaffSkill{
				AliasName: aliasName,
				Skills:    skills,
			}
			reception.StaffSkills = append(reception.StaffSkills, newStaffSkill)
		}
	}

	// **4. MongoDB に更新**
	update := bson.M{
		"$set": bson.M{
			"shifts": reception.Shifts,
			"work_staffs":  reception.WorkStaffs,
			"staff_skills":  reception.StaffSkills,
		},
	}

	updateResult, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Print(err, " Update error")
		http.Error(w, "Failed to update reception", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated %d document(s)", updateResult.ModifiedCount)


  responseData := struct {
    Csrf         string        `json:"csrf"`
    PushContents []string `json:"pushContents"`
    Reception collection.ReceptionStruct `json:"reception"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Reception: reception,
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

// `WorkStaffs` から `ShiftStart` に対応する `shift.Open` の AliasNames を取得
// func getOpenAliases(workStaffs []WorkStaff, shiftStart string, open int) []string {
//   var openAliases []string
//   count := 0
//   for _, staff := range workStaffs {
//     if staff.WorkStart == shiftStart {
//       openAliases = append(openAliases, staff.AliasName)
//       count++
//       if count >= open {
//         break
//       }
//     }
//   }
//   return openAliases
// }

// `shift.AliasNames` から `shift.Open` の範囲内のデータを取得
func getOpenAliasesFromShift(aliasNames []string, open int) []string {
  if len(aliasNames) > open {
    return aliasNames[:open]
  }
  return aliasNames
}

