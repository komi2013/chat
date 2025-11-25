package controller

import (
  "context"
  "encoding/json"
  // "fmt"
  "log"
  "net/http"
  // "strconv"
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
	receptionID := r.FormValue("receptionID")

  code := r.FormValue("code")
  if code == "" {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "code is invalid", http.StatusOK)
    return
  }

  // receptionOrders を受け取る
  var orders []ReceptionOrderStruct
  if err := json.Unmarshal([]byte(r.FormValue("receptionOrders")), &orders); err != nil {
      common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "JSON receptionOrders is wrong", http.StatusOK)
      return
  }

	// 必須チェック
	for _, order := range orders {
	    if order.ReceptionOrderID == "" ||
	        order.SeatName == "" ||
	        order.MenuID == 0 ||
	        order.Price == 0 {
	        common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "data is not enough", http.StatusOK)
	        return
	    }
	}

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
	collReception := common.DB.ReceptionDB.Collection("reception")
  var reception collection.ReceptionStruct
  filter := bson.M{"_id": receptionID}
  err = collReception.FindOne(ctx, filter).Decode(&reception)
  if err != nil {
    if err == mongo.ErrNoDocuments {
      common.WriteResponseWithSession(w, session, "Reception not found", http.StatusOK)
    } else {
      common.WriteResponseWithSession(w, session, "collReception.FindOne query failed", http.StatusOK)
    }
    return
  }

	collMenu := common.DB.ReceptionDB.Collection("menu")
  var menu collection.MenuStruct
  // filter := bson.M{"_id": receptionID}
  err = collMenu.FindOne(ctx, filter).Decode(&menu)
  if err != nil {
    common.WriteResponseWithSession(w, session, "collMenu.FindOne query failed", http.StatusOK)
    return
  }


	validCode := false
	seatName := ""
	for _, seat := range reception.Facilities {
		if seat.CurrentCode == code {
			seatName = seat.FacilityName
			validCode = true
			break
		}
	}
	if !validCode {
		common.WriteResponseWithSession(w, session, "コードが一致してません", http.StatusOK)
		return
	}

	orderUserIDs := append(reception.OrderUserIDs, session.UserID)
  collSessions := common.DB.SessionDB.Collection("session")
	filter = bson.M{"userID": bson.M{"$in": orderUserIDs},}
  cursor, err := collSessions.Find(context.TODO(), filter)
  if err != nil {
		common.WriteResponseWithSession(w, session, "collSessions.Find:", http.StatusOK)
		return
  }
  var sessions []collection.SessionStruct
  if err = cursor.All(context.TODO(), &sessions); err != nil {
		common.WriteResponseWithSession(w, session, "collSessions All:", http.StatusOK)
		return
  }

  for _, order := range orders {
    // log.Printf("受注: %+v\n", order)
    var selectedMenu *collection.Menu
    for _, menu := range menu.Menus {
      if menu.MenuID == order.MenuID {
        selectedMenu = &menu
        break
      }
    }
    if selectedMenu == nil {
      // log.Printf("MenuID %d not found", order.MenuID)
      continue
    }
    paidOptionSum := 0
    for _, opt := range order.PaidOptions {
      paidOptionSum += opt.Price
    }
    // calcTotal := selectedMenu.Price + paidOptionSum
    // if calcTotal != order.TotalPrice {
    //   log.Printf("⚠️ 金額差異: client=%d server=%d", order.TotalPrice, calcTotal)
    // }
    order.SeatName = seatName
  }

  var arr []interface{}
	arr = append(arr, "receptionOrder")
	arr = append(arr, reception.ChannelID)
	arr = append(arr, session.Nickname)
	arr = append(arr, orders)

// anybody any phone can order but order history will be different data 
// former phone order and latter phone order

	common.ChunkPush(sessions, arr)

	responseData := common.BaseResponse{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
	}
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

type ReceptionOrderStruct struct {
  ReceptionOrderID string            `json:"receptionOrderID"`
  SeatName         string            `json:"seatName"`
  MenuID           int               `json:"menuID"`
  MenuName         string            `json:"menuName"`
  ItemChoices      map[string]string `json:"itemChoices"`
  FreeOptions      []int             `json:"freeOptions"`
  PaidOptions      []struct {
      ItemID int `json:"itemID"`
      Price  int `json:"price"`
  } `json:"paidOptions"`
  FreeMultiOptions []int `json:"freeMultiOptions"`
  Price            int   `json:"price"`
  TotalPrice       int   `json:"totalPrice"`
}

