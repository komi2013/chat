package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  // "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionShift(w http.ResponseWriter, r *http.Request) {

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  receptionID := r.FormValue("receptionID")

	// var skills []string
	// if err := json.Unmarshal([]byte(r.FormValue("availableSkills")), &skills); err != nil {
	// 	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "availableSkills JSON Unmarshal Error", http.StatusOK)
	// 	return
	// }

	var updatedShifts []collection.Shift
	if err := json.Unmarshal([]byte(r.FormValue("updatedShifts")), &updatedShifts); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "updatedShifts JSON Unmarshal Error", http.StatusOK)
		return
	}

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
    }
  }
  if !staffAccess {
		common.WriteResponseWithSession(w, session, "!staffAccess:", http.StatusOK)
    return
  }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
	coll := common.DB.ReceptionDB.Collection("reception")

  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
		common.WriteResponseWithSession(w, session, "reception not found", http.StatusOK)
    return
  }

	for _, shift := range updatedShifts {
    existingStaffMap := make(map[string]bool)
    for _, staff := range reception.WorkStaffs {
      if staff.WorkStart == shift.ShiftStart {
        existingStaffMap[staff.AliasName] = true
      }
    }

    newAliasMap := make(map[string]bool)
    for _, alias := range shift.AliasNames {
      newAliasMap[alias] = true
    }

    newOpenAliases := getOpenAliasesFromShift(shift.AliasNames, shift.Open)

    var addedAliases []string
    for i, alias := range shift.AliasNames {
      if i < shift.Open && !existingStaffMap[alias] {
        addedAliases = append(addedAliases, alias)
      }
    }

    var removedAliases []string
    for alias := range existingStaffMap {
      if !newAliasMap[alias] || !common.SliceStrContains(newOpenAliases, alias) {
        removedAliases = append(removedAliases, alias)
      }
    }

    for _, aliasName := range addedAliases {
      newStaff := collection.WorkStaff{
        AliasName: aliasName,
        WorkStart: shift.ShiftStart,
        WorkEnd:   shift.ShiftEnd,
        Seq:       len(reception.WorkStaffs) + 1,
      }
      reception.WorkStaffs = append(reception.WorkStaffs, newStaff)
    }

    for j := 0; j < len(reception.WorkStaffs); {
      staff := reception.WorkStaffs[j]
      if common.SliceStrContains(removedAliases, staff.AliasName) && staff.WorkStart == shift.ShiftStart {
        reception.WorkStaffs = append(reception.WorkStaffs[:j], reception.WorkStaffs[j+1:]...)
      } else {
        j++
      }
    }

		// log.Printf("ShiftStart: %s", shift.ShiftStart)
		// log.Printf("Added AliasNames: %v", addedAliases)
		// log.Printf("Removed AliasNames: %v", removedAliases)

		for i, s := range reception.Shifts {
			if s.ShiftStart == shift.ShiftStart && s.Skill == shift.Skill {
				reception.Shifts[i].AliasNames = shift.AliasNames
				reception.Shifts[i].Fix = shift.Fix
			}
		}
	}

	// log.Printf("Updated WorkStaffs: %+v", reception.WorkStaffs)
	var receptionLog = common.NewDailyLogger("reception_")
	// if len(skills) > 0 {
	// 	updated := false
	// 	for j, staffSkill := range reception.StaffSkills {
	// 		if staffSkill.AliasName == aliasName {
	// 			reception.StaffSkills[j].Skills = skills
	// 			updated = true
	// 			break
	// 		}
	// 	}

	// 	if !updated {
	// 		newStaffSkill := collection.StaffSkill{
	// 			AliasName: aliasName,
	// 			Skills:    skills,
	// 		}
	// 		reception.StaffSkills = append(reception.StaffSkills, newStaffSkill)
	// 	}
	// }

	// 過去のShiftEndを持つシフトを削除
	now := time.Now()
	var validShifts []collection.Shift
	for _, shift := range reception.Shifts {
		// ShiftEndをパース（datetime-local形式: "2006-01-02T15:04"）
		shiftEndTime, err := time.Parse("2006-01-02T15:04", shift.ShiftEnd)
		if err != nil {
			// パースエラーの場合は保持（形式が異なる可能性があるため）
			validShifts = append(validShifts, shift)
			continue
		}

		// 現在時刻より未来のシフトのみ保持
		if shiftEndTime.After(now) {
			validShifts = append(validShifts, shift)
		}
	}

	// 過去のシフトを除外したリストで更新
	reception.Shifts = validShifts

	update := bson.M{
		"$set": bson.M{
			"shifts": reception.Shifts,
			"workStaffs":  reception.WorkStaffs,
			// "staffSkills":  reception.StaffSkills,
		},
	}

	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		receptionLog.Printf("updateResult coll.UpdateOne: %v; Req: ", err, r.URL.Path, r.Form)
		common.WriteResponseWithSession(w, session, "reception not found", http.StatusOK)
		return
	}

	// log.Printf("Updated %d document(s)", updateResult.ModifiedCount)

	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception,
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

