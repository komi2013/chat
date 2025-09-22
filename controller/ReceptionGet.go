package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "time"

  // "go.mongodb.org/mongo-driver/mongo"
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
	switch codeType {
	case "1":  // before enter
	case "3":  // open menu
	default:
    codeType = "2" // order at seat
	}

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

	// after you get reception DB, more strict check
	staffAccess = false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID && channelID == reception.ChannelID {
      staffAccess = true
    }
  }
  var receptionLog = common.NewDailyLogger("reception_")

  seatName := ""
	validCode := false
	if codeType == "1" {
		now := time.Now()
    for j, passcode := range reception.Passcodes {
			receptionLog.Printf("passcode.Passkey == code:", passcode.Passkey, code)
      if passcode.Passkey == code {
				startTime, err1 := time.Parse("2006-01-02T15:04", passcode.PassStart)
				endTime, err2 := time.Parse("2006-01-02T15:04", passcode.PassEnd)
	      if err1 == nil && err2 == nil && now.After(startTime) && now.Before(endTime) {
	        validCode = true
	        break
	      } else if now.After(endTime) {
	        reception.Passcodes = append(
	            reception.Passcodes[:j],
	            reception.Passcodes[j+1:]...,
	        )
	        update := bson.M{
	            "$set": bson.M{
	                "passcodes": reception.Passcodes,
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
    }
	} else if codeType == "2" {
		for _, seat := range reception.Seats {
			if seat.CurrentCode == code {
				validCode = true
				seatName = seat.SeatName
				break
			}
			if validCode {
				break
			}
		}
	} else if codeType == "3" {
	    for i, seat := range reception.Seats {
	        for j, passcode := range seat.Passcodes {
	            if passcode.Passkey == code {
	                validCode = true
	                reception.Seats[i].CurrentCode = code
	                reception.Seats[i].Passcodes = append(
	                    seat.Passcodes[:j],
	                    seat.Passcodes[j+1:]...,
	                )
	                update := bson.M{
	                    "$set": bson.M{
	                        "seats": reception.Seats,
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
	}

  if !staffAccess && !validCode{
	  if !validCode {
	  	receptionLog.Printf("validCode is invalid", r.URL.Path, r.Form)
	  	common.WriteResponseWithSession(w, session, "コードが一致してません", http.StatusOK)
	  	return
	  }
	  if !staffAccess {
	    receptionLog.Printf("ChannelAliases !staffAccess: %v; Req: ", session.ChannelAliases, r.URL.Path, r.Form)
	    common.WriteResponseWithSession(w, session, "コードが一致してません", http.StatusOK)
	    return
	  }
  }

	if !staffAccess {
		reception.ChannelID = ""
	  reception.AdminNames = nil
	  reception.Passcodes = nil
	  reception.JoinNames = nil
	  for i := range reception.Seats {
	    reception.Seats[i].Passcodes = nil
	  }
	}

  // reception の値だけここでセット
	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Mail:         session.Mail,
		Telephone:    session.Telephone,
		Reception:    reception,
		SeatName:     seatName,
		Nickname:     session.Nickname,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
