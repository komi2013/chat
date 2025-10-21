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
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  "chat/collection"
  "chat/common"
)

func ReceptionGet(w http.ResponseWriter, r *http.Request) {

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")
  receptionID := r.FormValue("receptionID")
  code := r.FormValue("code")
  codeType := r.FormValue("codeType")

  // 1 = at reception, 2 = order, 3 = at seat
	// if codeType != "1" && codeType != "2" && codeType != "3" {
	// 	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid codeType", http.StatusOK)
	// 	return
	// }

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

  // if customer use own channelID, you can not block 100%
  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
    }
  }

	if !staffAccess && code == "" {
		common.WriteResponseWithSession(w, session, "コードがない", http.StatusOK)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  var reception collection.ReceptionStruct
	coll := common.DB.ReceptionDB.Collection("reception")
	filter := bson.M{"_id": receptionID}
	err = coll.FindOne(ctx, filter).Decode(&reception)
	if err != nil {
	  common.WriteResponseWithSession(w, session, err.Error(), http.StatusOK)
		return
	}
  var receptionLog = common.NewDailyLogger("reception_")

	// after you get reception DB, more strict check
	staffAccess = false
  for _, d := range session.ChannelAliases {
  	receptionLog.Printf("d.Alias == aliasName && d.ChannelID == channelID && channelID == reception.ChannelID", d.Alias, aliasName, d.ChannelID, channelID, reception.ChannelID)
    if d.Alias == aliasName && d.ChannelID == channelID && channelID == reception.ChannelID {
      staffAccess = true
    }
  }

  seatName := ""
	validCode := false
	if codeType == "2" {
	    for i, seat := range reception.Facilities {
	        for j, passcode := range seat.Passcodes {
	            if passcode.Passkey == code {
	                validCode = true
	                reception.Facilities[i].CurrentCode = code
	                reception.Facilities[i].Passcodes = append(
	                    seat.Passcodes[:j],
	                    seat.Passcodes[j+1:]...,
	                )
	                update := bson.M{
	                    "$set": bson.M{
	                        "facilities": reception.Facilities,
	                    },
	                }
	                _, err := coll.UpdateOne(ctx, filter, update)
	                if err != nil {
	                    common.WriteResponseWithSession(w, session, "Update failed: "+err.Error(), http.StatusInternalServerError)
	                    return
	                }

	                break
	            }
	        }
	        if validCode {
	            break
	        }
	    }
	} else if codeType == "3" {
		for _, seat := range reception.Facilities {
			if seat.CurrentCode == code {
				// receptionLog.Printf("seat.CurrentCode == code", seat.CurrentCode, code)
				validCode = true
				seatName = seat.FacilityName
				break
			}
		}
	}
	receptionLog.Printf("seatName", seatName)
  if !staffAccess && !validCode && codeType != "" {
  	common.WriteResponseWithSession(w, session, "コードが一致してません。スタッフでもありません", http.StatusOK)
  	return
  }

	collMenu := common.DB.ReceptionDB.Collection("menu")
  var menu collection.MenuStruct
  err = collMenu.FindOne(ctx, filter).Decode(&menu)
  if err != nil && err != mongo.ErrNoDocuments {
    common.WriteResponseWithSession(w, session, "collMenu.FindOne query failed"+err.Error(), http.StatusOK)
    return
  }

	if !staffAccess {
		reception.ChannelID = ""
	  reception.AdminNames = nil
	  reception.Passcodes = nil
	  reception.JoinNames = nil
	  for i := range reception.Facilities {
	    reception.Facilities[i].Passcodes = nil
	  }
	}

  // reception の値だけここでセット
	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Mail:         session.Mail,
		Telephone:    session.Telephone,
		Reception:    reception,
		Menu:    			menu,
		FacilityName: seatName,
		Nickname:     session.Nickname,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
