package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  "strconv"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  "chat/collection"
  "chat/common"
)

func ReceptionOrder(w http.ResponseWriter, r *http.Request) {
	postBy := r.FormValue("postBy")
	receptionID := r.FormValue("receptionID")
  menuID, err := strconv.Atoi(r.FormValue("menuID"))
  if err != nil {
    http.Error(w, "menuID is not correct", http.StatusNotFound)
    return
  }

  code := r.FormValue("code")
  if code == "" {
    http.Error(w, "Code is required", http.StatusBadRequest)
    return
  }

  // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  // defer cancel()
  // c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  // if err != nil {
  //   log.Printf("mongo.Connect: %v; Req: ", err, r.URL.Path, r.Form)
  // }
  // defer c.Disconnect(ctx)
  // db1 := c.Database(common.MongoDb1)

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    log.Printf("SessionCheck: %v; Req: ", err, r.URL.Path, r.Form)
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }

	trueAccess := false
	for _, d := range session.ChannelAliases {
		if d.Alias == postBy {
			trueAccess = true
			break
		}
	}
	if !trueAccess {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
	coll := common.DB.ReceptionDB.Collection("reception")
  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = coll.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      http.Error(w, "Reception not found", http.StatusNotFound)
    } else {
      http.Error(w, "Database query failed", http.StatusInternalServerError)
    }
    return
  }

  found := false
  tableName := ""
  for i, table := range reception.Seats {
    for _, passcode := range table.Passcodes {
      if passcode.Passkey == code {
        found = true
        reception.Seats[i].CurrentCode = code
        tableName = table.SeatName
      }
    }
    if found {
      break
    }
  }

  if !found {
    http.Error(w, "Code not associated with any table", http.StatusNotFound)
    return
  }


	// Find the requested menu
	var selectedMenu *collection.Menu
	for _, menu := range reception.Menus {
		if menu.MenuID == menuID {
			selectedMenu = &menu
			break
		}
	}
	// if selectedMenu == nil {
	// 	return nil, fmt.Errorf("menuID %d not found", menuID)
	// }

	// if menuPrice == 0 {
	// 	http.Error(w, "Menu not found or price unavailable", http.StatusNotFound)
	// 	return
	// }

	var freeOptions []int
  if err := json.Unmarshal([]byte(r.FormValue("freeOptions")), &freeOptions); err != nil {
    http.Error(w, "Invalid JSON freeOptions", http.StatusBadRequest)
    return
  }

  var freeMultiOptions []int
  if err := json.Unmarshal([]byte(r.FormValue("freeMultiOptions")), &freeMultiOptions); err != nil {
    http.Error(w, "Invalid JSON freeMultiOptions", http.StatusBadRequest)
    return
  }

  var paidOptions []int
  if err := json.Unmarshal([]byte(r.FormValue("paidOptions")), &paidOptions); err != nil {
    http.Error(w, "Invalid JSON paidOptions", http.StatusBadRequest)
    return
  }

	var itemIDs []int
	var paidOptionSum int

	// FreeOptions (単純なintスライスに対応)
	for _, freeID := range freeOptions {
		for _, IDs := range selectedMenu.FreeOptions {
			for _, ID := range IDs {
				if freeID == ID {
					for _, item := range reception.ItemDetails {
						if item.ItemID == freeID {
							itemIDs = append(itemIDs, item.ItemID)
							break
						}
					}
				}
			}
		}
	}

	// FreeMultiOptions (intスライス)
	for _, freeID := range freeMultiOptions {
		for _, ID := range selectedMenu.FreeMultiOptions {
			if freeID == ID {
				for _, item := range reception.ItemDetails {
					if item.ItemID == freeID {
						itemIDs = append(itemIDs, item.ItemID)
						break
					}
				}
			}
		}
	}

	// PaidOptions (構造体: []ItemOption)
	for _, paidID := range paidOptions {
		for _, option := range selectedMenu.PaidOptions {
			if option.ItemID == paidID {
				for _, item := range reception.ItemDetails {
					if item.ItemID == paidID {
						itemIDs = append(itemIDs, item.ItemID)
						paidOptionSum += option.Price
						break
					}
				}
			}
		}
	}

	menuPrice := selectedMenu.Price + paidOptionSum

  var arr []interface{}
  arr = append(arr, "receptionOrder")
  arr = append(arr, tableName)
  arr = append(arr, r.FormValue("menuID"))
  arr = append(arr, itemIDs)
  arr = append(arr, menuPrice)
  arr = append(arr, common.StringRand(4))
  // for _, subscription := range reception.Subscriptions {
	 //  pushID := common.StringRand(12)
	 //  arrForPush := append([]interface{}{pushID}, arr...)
	  
	 //  resp, err := common.PushNotification(string(arrForPush), subscription)
  //   if err != nil {
	 //    log.Printf("PushNotification: %v; Req: ", err, r.URL.Path, r.Form)
	 //    http.Error(w, err.Error(), http.StatusServiceUnavailable)
  //   }
  //   defer resp.Body.Close()
  // }

	responseData := struct {
		Csrf         string   `json:"csrf"`
		PushContents []string `json:"pushContents"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)

}

