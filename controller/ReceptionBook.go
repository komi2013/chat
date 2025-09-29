package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/bson/primitive"

	"chat/collection"
	"chat/common"
)

func ReceptionBook(w http.ResponseWriter, r *http.Request) {
	menuID, err := strconv.Atoi(r.FormValue("menuID"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid menuID format", http.StatusOK)
		return
	}
	receptionID := r.FormValue("receptionID")
	var answers []string
	if err := json.Unmarshal([]byte(r.FormValue("answers")), &answers); err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid JSON answers", http.StatusOK)
		return
	}
	bookStart := r.FormValue("bookStart")
	bookEnd := r.FormValue("bookEnd")
	if bookStart == "" || bookEnd == "" {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Missing bookStart or bookEnd", http.StatusOK)
		return
	}
	layout := "2006-01-02T15:04"
	bookStartTime, err := time.Parse(layout, bookStart)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid bookStart format", http.StatusOK)
		return
	}
	bookEndTime, err := time.Parse(layout, bookEnd)
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid bookEnd format", http.StatusOK)
		return
	}

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		log.Printf("SessionCheckTake: %v; Req: ", err, r.URL.Path)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	trueAccess := false
	if session.Mail != "" && session.Telephone != "" {
		trueAccess = true
	}
	if !trueAccess {
		common.WriteResponseWithSession(w, session, "Unauthorized", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var reception collection.ReceptionStruct
	filter := bson.M{"_id": receptionID}

	coll := common.DB.ReceptionDB.Collection("reception")
	err = coll.FindOne(ctx, filter).Decode(&reception)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to find reception:" + err.Error(), http.StatusOK)
		return
	}

	var menu collection.MenuStruct
	collMenu := common.DB.ReceptionDB.Collection("menu")
	err = collMenu.FindOne(ctx, filter).Decode(&menu)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to find menu:" + err.Error(), http.StatusOK)
		return
	}

	var selectedService *collection.Menu
	for _, service := range menu.Menus {
		if service.MenuID == menuID {
			selectedService = &service
			break
		}
	}
	if selectedService == nil {
		common.WriteResponseWithSession(w, session, "Service not found", http.StatusOK)
		return
	}

	needSkill := selectedService.NeedSkill
	needFacility := selectedService.NeedFacility

	validStaffCount := 0
	for _, ws := range reception.WorkStaffs {
	    workStart, _ := time.Parse(layout, ws.WorkStart)
	    workEnd, _ := time.Parse(layout, ws.WorkEnd)
	    if !(bookStartTime.Before(workStart) || bookEndTime.After(workEnd)) {
	        if needSkill == "" || hasValidSkill(ws.AliasName, needSkill, reception.StaffSkills) {
	            validStaffCount++
	        }
	    }
	}

	if validStaffCount == 0 && reception.WorkStaffNeed {
		common.WriteResponseWithSession(w, session, "No available staff", http.StatusOK)
		return
	}

	if needFacility != "" && !isFacilityAvailable(bookStartTime, bookEndTime, needFacility, reception, menu) {
		common.WriteResponseWithSession(w, session, "Facility not available", http.StatusOK)
		return
	}

	overlappingCount := 0
	for _, booking := range reception.Books {
		bookedStart, _ := time.Parse(layout, booking.BookStart)
		bookedEnd, _ := time.Parse(layout, booking.BookEnd)
		if bookStartTime.Before(bookedEnd) && bookEndTime.After(bookedStart) {
			overlappingCount++
		}
	}

	if overlappingCount >= validStaffCount && reception.WorkStaffNeed{
		common.WriteResponseWithSession(w, session, "this time already booked", http.StatusOK)
		return
	}

	newBooking := collection.Book{
		BookStart: bookStartTime.Format(layout),
		BookEnd:   bookEndTime.Format(layout),
		Answers:   answers,
		MenuID: menuID,
		CreatedAt: time.Now(),
		Nickname: session.Nickname,
	}
	reception.Books = append(reception.Books, newBooking)

	update := bson.M{
		"$set": bson.M{"books": reception.Books},
	}
	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to update booking", http.StatusOK)
		return
	}

	responseData := common.BaseResponse{
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

func isFacilityAvailable(start, end time.Time, facilityName string, reception collection.ReceptionStruct, menu collection.MenuStruct) bool {
	layout := "2006-01-02T15:04"

	// 必要な設備の情報を取得
	var facilityCount int
	for _, facility := range reception.Facilities {
		if facility.Bookable && facility.FacilityName == facilityName {
			facilityCount = facility.Capacity
			break
		}
	}

	// 設備が登録されていない場合は予約不可
	if facilityCount == 0 {
		return false
	}

	// 指定された設備を利用する予約の数をカウント
	overlappingBookings := 0
	for _, book := range reception.Books {
		bookedStart, _ := time.Parse(layout, book.BookStart)
		bookedEnd, _ := time.Parse(layout, book.BookEnd)

		// 予約時間が重なっているかチェック
		if start.Before(bookedEnd) && end.After(bookedStart) {
			// book.ServiceID に対応する Service を取得し、必要な設備を確認
			for _, service := range menu.Menus {
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
