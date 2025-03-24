package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"chat/collection"
	"chat/common"
)

func BookAdd(w http.ResponseWriter, r *http.Request) {
	postBy := r.FormValue("postBy")
	bookStart := r.FormValue("bookStart")
	bookEnd := r.FormValue("bookEnd")
	menuID, err := strconv.Atoi(r.FormValue("menuID"))
	if err != nil {
		http.Error(w, "Invalid menuID format", http.StatusBadRequest)
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

	layout := "2006-01-02T15:04"
	bookStartTime, err := time.Parse(layout, bookStart)
	if err != nil {
		http.Error(w, "Invalid bookStart format", http.StatusBadRequest)
		return
	}
	bookEndTime, err := time.Parse(layout, bookEnd)
	if err != nil {
		http.Error(w, "Invalid bookEnd format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Printf("mongo.Connect error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer c.Disconnect(ctx)

	db1 := c.Database(common.MongoDb1)

	session, err := common.SessionCheck(db1, w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheck error: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 投稿者が正しいかチェック
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

	// bookPattern を取得
	var bookPattern collection.BookPatternStruct
	filter := bson.M{"_id": bookPatternID}
	coll := db1.Collection("book_pattern")

	err = coll.FindOne(ctx, filter).Decode(&bookPattern)
	if err != nil {
		log.Printf("Failed to find bookPattern: %v", err)
		http.Error(w, "BookPattern not found", http.StatusNotFound)
		return
	}

	// 選択したサービスを取得
	var selectedService *collection.Menu
	for _, service := range bookPattern.Menus {
		if service.MenuID == menuID {
			selectedService = &service
			break
		}
	}
	if selectedService == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	needSkill := selectedService.NeedSkill
	needFacility := selectedService.NeedFacility

	validStaffCount := 0
	for _, ws := range bookPattern.WorkStaffs {
	    workStart, _ := time.Parse(layout, ws.WorkStart)
	    workEnd, _ := time.Parse(layout, ws.WorkEnd)

	    log.Printf("Checking workStaff: %s (WorkStart: %s, WorkEnd: %s)", ws.AliasName, ws.WorkStart, ws.WorkEnd)

	    // 予約時間がスタッフの勤務時間内にあるか
	    if !(bookStartTime.Before(workStart) || bookEndTime.After(workEnd)) {
	        log.Printf("Staff %s is within work hours", ws.AliasName)

	        // スキルチェック（needSkill が設定されている場合のみ）
	        if needSkill == "" || hasValidSkill(ws.AliasName, needSkill, bookPattern.StaffSkills) {
	            log.Printf("Staff %s has the required skill", ws.AliasName)
	            validStaffCount++
	        } else {
	            log.Printf("Staff %s does NOT have the required skill (%s)", ws.AliasName, needSkill)
	        }
	    } else {
	        log.Printf("Staff %s is NOT available in the requested time slot", ws.AliasName)
	    }
	}

	// スタッフがいなければ予約不可
	if validStaffCount == 0 {
	    log.Printf("No valid staff found for this booking")
	    http.Error(w, "No available staff", http.StatusConflict)
	    return
	}

	// 設備の空き状況をチェック
	if needFacility != "" && !isFacilityAvailable(bookStartTime, bookEndTime, needFacility, bookPattern) {
		http.Error(w, "Facility not available", http.StatusConflict)
		return
	}

	// 予約の競合をチェック（複数予約を考慮）
	overlappingCount := 0
	for _, booking := range bookPattern.Books {
		bookedStart, _ := time.Parse(layout, booking.BookStart)
		bookedEnd, _ := time.Parse(layout, booking.BookEnd)

		if bookStartTime.Before(bookedEnd) && bookEndTime.After(bookedStart) {
			overlappingCount++
		}
	}

	// 予約枠が埋まっていないかチェック（workStaff数と比較）
	if overlappingCount >= validStaffCount {
		http.Error(w, "Time slot already booked", http.StatusConflict)
		return
	}

	// 予約を追加
	newBooking := collection.Book{
		BookStart: bookStartTime.Format(layout),
		BookEnd:   bookEndTime.Format(layout),
		Answers:   answers,
		MenuID: menuID,
		CreatedAt: time.Now(),
	}
	bookPattern.Books = append(bookPattern.Books, newBooking)

	// DB更新
	update := bson.M{
		"$set": bson.M{"books": bookPattern.Books},
	}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update booking", http.StatusInternalServerError)
		return
	}

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

// hasValidSkill スタッフがスキルを持っているか確認
func hasValidSkill(aliasName, needSkill string, staffSkills []collection.StaffSkill) bool {
	for _, staff := range staffSkills {
		if staff.AliasName == aliasName {
			for _, skill := range staff.Skills {
				if skill == needSkill {
					return true
				}
			}
		}
	}
	return false
}

func isFacilityAvailable(start, end time.Time, facilityName string, bookPattern collection.BookPatternStruct) bool {
	layout := "2006-01-02T15:04"

	// 必要な設備の情報を取得
	var facilityCount int
	for _, facility := range bookPattern.Facilities {
		if facility.FacilityName == facilityName {
			facilityCount = facility.FacilityCount
			break
		}
	}

	// 設備が登録されていない場合は予約不可
	if facilityCount == 0 {
		return false
	}

	// 指定された設備を利用する予約の数をカウント
	overlappingBookings := 0
	for _, book := range bookPattern.Books {
		bookedStart, _ := time.Parse(layout, book.BookStart)
		bookedEnd, _ := time.Parse(layout, book.BookEnd)

		// 予約時間が重なっているかチェック
		if start.Before(bookedEnd) && end.After(bookedStart) {
			// book.ServiceID に対応する Service を取得し、必要な設備を確認
			for _, service := range bookPattern.Menus {
				if service.MenuID == book.MenuID && service.NeedFacility == facilityName {
					overlappingBookings++
					break
				}
			}
		}
	}

	// 予約が設備の数より少なければ予約可能
	return overlappingBookings < facilityCount
}
