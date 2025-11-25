package controller

import (
  "context"
  "encoding/json"
  // "io"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/collection"
  "chat/common"
)

func ReceptionEdit(w http.ResponseWriter, r *http.Request) {

	// log.Printf("r.FormValue: %v", r.FormValue("reception"))
  var reception collection.ReceptionStruct
  if err := json.Unmarshal([]byte(r.FormValue("reception")), &reception); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON reception", http.StatusOK)
		return
  }

  var menu collection.MenuStruct
  if err := json.Unmarshal([]byte(r.FormValue("menu")), &menu); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON menu", http.StatusOK)
		return
  }

  aliasName := r.FormValue("aliasName")
  channelID := r.FormValue("channelID")

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  staffAccess := false
  for _, d := range session.ChannelAliases {
    if d.Alias == aliasName && d.ChannelID == channelID {
      staffAccess = true
      break
    }
  }

  if !staffAccess {
		common.WriteResponseWithSession(w, session, "no true staff access right", http.StatusOK)
		return
  }

	for i := range menu.ItemDetails {
    if menu.ItemDetails[i].ImgPath != "" {
      savedPath, err := common.ImgSave(
        menu.ItemDetails[i].ImgPath,
        session.UserID,
        aliasName,
        channelID,
        3,
        6,
      )
      if err != nil {
        common.WriteResponseWithSession(w, session, "ItemDetails[i].ImgPath: " + err.Error(), http.StatusOK)
        return
      }
      menu.ItemDetails[i].ImgPath = savedPath
    }
	}
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  coll := common.DB.ReceptionDB.Collection("reception")
	var existing collection.ReceptionStruct
	filter := bson.M{"_id": reception.ReceptionID}
	err = coll.FindOne(ctx, filter).Decode(&existing)

	if err != nil && err != mongo.ErrNoDocuments {
		common.WriteResponseWithSession(w, session, "reception FindOne: " + err.Error(), http.StatusOK)
		return
	}
	adminAccess := false
	for _, admin := range existing.AdminNames {
		if admin == aliasName {
			adminAccess = true
			break
		}
	}
	if !adminAccess && !staffAccess {
		var receptionLog = common.NewDailyLogger("reception_")
		receptionLog.Printf("Unauthorized edit attempt: aliasName=%s, channelID=%s, receptionID=%s", aliasName, channelID, reception.ReceptionID)
		common.WriteResponseWithSession(w, session, "You do not have permission to edit this reception", http.StatusOK)
		return
	}

	filter = bson.M{"_id": reception.ReceptionID}
	reception.UpdatedAt = time.Now()
	data, err := bson.Marshal(reception)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to marshal reception", http.StatusOK)    
    return
	}

	var updateData bson.M
	if err := bson.Unmarshal(data, &updateData); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to unmarshal reception to bson.M", http.StatusOK)
    return
	}
	delete(updateData, "_id")
	update := bson.M{"$set": updateData}
	opts := options.Update().SetUpsert(true)
	_, err = coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
    common.WriteResponseWithSession(w, session, "Failed to update reception", http.StatusOK)
    return
	}

	// ---- menu 更新 ----
	menuColl := common.DB.ReceptionDB.Collection("menu")
	menuFilter := bson.M{"_id": reception.ReceptionID}

	// menu に UpdatedAt を入れる
	menu.UpdatedAt = time.Now()

	menuData, err := bson.Marshal(menu)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to marshal menu", http.StatusOK)
		return
	}

	var menuUpdateData bson.M
	if err := bson.Unmarshal(menuData, &menuUpdateData); err != nil {
		common.WriteResponseWithSession(w, session, "Failed to unmarshal menu to bson.M", http.StatusOK)
		return
	}
	delete(menuUpdateData, "_id")

	menuUpdate := bson.M{"$set": menuUpdateData}
	menuOpts := options.Update().SetUpsert(true)
	_, err = menuColl.UpdateOne(ctx, menuFilter, menuUpdate, menuOpts)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to update menu", http.StatusOK)
		return
	}

	responseData := common.ReceptionResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Reception:    reception,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}
