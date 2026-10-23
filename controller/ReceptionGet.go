package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/common"
	"chat/table"
)

// receptionQueryer is implemented by both *sql.DB and *sql.Tx.
type receptionQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type receptionMenuResponse struct {
	ReceptionID int64              `json:"receptionID"`
	UpdatedAt   time.Time          `json:"updatedAt,omitempty"`
	Menus       []table.Menu       `json:"menus"`
	ItemDetails []table.ItemDetail `json:"itemDetails"`
}

type receptionBookResponse struct {
	BookID    int64     `json:"bookID"`
	BookStart string    `json:"bookStart"`
	BookEnd   string    `json:"bookEnd"`
	Answers   []string  `json:"answers,omitempty"`
	MenuID    int64     `json:"menuID,omitempty"`
	People    int       `json:"people,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type receptionQueueResponse struct {
	WaitingGuest int    `json:"waitingGuest,omitempty"`
	QueueName    string `json:"queueName,omitempty"`
	QueuedAt     string `json:"queuedAt,omitempty"`
	UserID       int64  `json:"userID,omitempty"`
}

// receptionGetResponse is the public reception shape. Menu master data is
// returned in the separate top-level "menu" property, not duplicated here.
type receptionGetResponse struct {
	ReceptionID     int64                    `json:"receptionID"`
	ChannelID       int64                    `json:"channelID"`
	AdminNames      []string                 `json:"adminNames"`
	JoinNames       []string                 `json:"joinNames"`
	InquiryNames    []string                 `json:"inquiryNames,omitempty"`
	OrderUserIDs    []int64                  `json:"orderUserIDs,omitempty"`
	UpdatedAt       time.Time                `json:"updatedAt,omitempty"`
	ReceptionTitle  string                   `json:"receptionTitle,omitempty"`
	Skills          []string                 `json:"skills,omitempty"`
	WorkStaffNeed   bool                     `json:"workStaffNeed"`
	Passcodes       []table.Passcode         `json:"passcodes"`
	Asks            []table.Ask              `json:"asks"`
	AskChoices      []table.AskChoice        `json:"askChoices"`
	AskMultiChoices []table.AskMultiChoice   `json:"askMultiChoices"`
	Facilities      []table.FacilityView     `json:"facilities"`
	OpenTimes       []table.OpenTime         `json:"openTimes"`
	Shifts          []table.Shift            `json:"shifts"`
	StaffSkills     []table.StaffSkill       `json:"staffSkills"`
	WorkStaffs      []table.WorkStaff        `json:"workStaffs"`
	WaitConfigs     []table.WaitConfig       `json:"waitConfigs"`
	Books           []receptionBookResponse  `json:"books"`
	Queues          []receptionQueueResponse `json:"queues"`
}

func ReceptionGet(w http.ResponseWriter, r *http.Request) {
	aliasName := r.FormValue("aliasName")
	channelIDText := r.FormValue("channelID") // Optional for public reception/booking views.
	receptionID, err := strconv.ParseInt(r.FormValue("receptionID"), 10, 64)
	if err != nil || receptionID <= 0 {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid receptionID", http.StatusOK)
		return
	}

	var requestedChannelID int64
	if channelIDText != "" {
		requestedChannelID, err = strconv.ParseInt(channelIDText, 10, 64)
		if err != nil || requestedChannelID <= 0 {
			common.WriteResponseWithoutSession(w, r.FormValue("csrf"), "Invalid channelID", http.StatusOK)
			return
		}
	}

	code := r.FormValue("code")
	codeType := r.FormValue("codeType")

	session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
	if err != nil {
		common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	db := common.DB.ReceptionDB
	baseReception, err := loadReceptionBase(ctx, db, receptionID)
	if err != nil {
		message := err.Error()
		if err == sql.ErrNoRows {
			message = "Reception not found"
		}
		common.WriteResponseWithSession(w, session, message, http.StatusOK)
		return
	}

	// A staff match requires the authenticated user, alias name, and reception channel.
	staffAccess := false
	if aliasName != "" && requestedChannelID > 0 && requestedChannelID == baseReception.ChannelID {
		var aliasExists bool
		err = common.DB.AliasDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM alias
				WHERE user_id = $1 AND alias_name = $2 AND channel_id = $3
			)
		`, session.UserID, aliasName, requestedChannelID).Scan(&aliasExists)
		if err != nil {
			common.WriteResponseWithSession(w, session, "Failed to validate alias: "+err.Error(), http.StatusOK)
			return
		}
		staffAccess = aliasExists
	}

	seatName := ""
	validCode := false
	switch codeType {
	case "":
		// Public mode intentionally does not require a code.
	case "2":
		if code != "" {
			validCode, seatName, err = consumeReceptionFacilityPasscode(ctx, db, receptionID, code)
			if err != nil {
				common.WriteResponseWithSession(w, session, "Passcode update failed: "+err.Error(), http.StatusOK)
				return
			}
		}
	case "3":
		if code != "" {
			err = db.QueryRowContext(ctx, `
				SELECT COALESCE(f.facility->>'facilityName', '')
				FROM reception r
				CROSS JOIN LATERAL jsonb_array_elements(r.facilities) AS f(facility)
				JOIN reception_facility_code c
				  ON c.reception_id = r.reception_id
				 AND c.facility_id = (f.facility->>'facilityID')::BIGINT
				WHERE r.reception_id = $1 AND c.current_code = $2
				LIMIT 1
			`, receptionID, code).Scan(&seatName)
			if err == nil {
				validCode = true
			} else if err != sql.ErrNoRows {
				common.WriteResponseWithSession(w, session, "Failed to validate facility code: "+err.Error(), http.StatusOK)
				return
			}
		}
	case "1":
		// Not implemented: this code type cannot bypass validation for non-staff.
	default:
		// Unknown code types are invalid for non-staff users.
	}

	if !staffAccess && codeType != "" && !validCode {
		common.WriteResponseWithSession(w, session, "コードが一致していません。スタッフでもありません", http.StatusOK)
		return
	}

	reception, err := loadReception(ctx, db, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		return
	}
	menu := receptionMenuResponse{
		ReceptionID: reception.ReceptionID,
		UpdatedAt:   reception.UpdatedAt,
		Menus:       reception.Menus,
		ItemDetails: reception.ItemDetails,
	}

	facilities, err := loadReceptionFacilityViews(ctx, db, receptionID, reception.Facilities)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load facilities: "+err.Error(), http.StatusOK)
		return
	}
	books, err := loadReceptionBooks(ctx, db, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load bookings: "+err.Error(), http.StatusOK)
		return
	}
	queues, err := loadReceptionQueues(ctx, db, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load queues: "+err.Error(), http.StatusOK)
		return
	}

	if !staffAccess {
		reception.ChannelID = 0
		reception.AdminNames = nil
		reception.Passcodes = nil
		reception.JoinNames = nil
		reception.InquiryNames = nil
		reception.OrderUserIDs = nil
		for i := range facilities {
			facilities[i].Passcodes = nil
		}
	}

	responseData := struct {
		Csrf         string                `json:"csrf"`
		PushContents []string              `json:"pushContents"`
		Mail         string                `json:"mail,omitempty"`
		Telephone    string                `json:"telephone,omitempty"`
		UserID       int64                 `json:"userID"`
		Reception    receptionGetResponse  `json:"reception"`
		Menu         receptionMenuResponse `json:"menu"`
		FacilityName string                `json:"facilityName"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Mail:         session.Mail,
		Telephone:    session.Telephone,
		UserID:       session.UserID,
		Reception: receptionGetResponse{
			ReceptionID: reception.ReceptionID, ChannelID: reception.ChannelID,
			AdminNames: reception.AdminNames, JoinNames: reception.JoinNames,
			InquiryNames: reception.InquiryNames,
			OrderUserIDs: reception.OrderUserIDs, UpdatedAt: reception.UpdatedAt,
			ReceptionTitle: reception.ReceptionTitle, Skills: reception.Skills,
			WorkStaffNeed: reception.WorkStaffNeed, Passcodes: reception.Passcodes,
			Asks: reception.Asks, AskChoices: reception.AskChoices,
			AskMultiChoices: reception.AskMultiChoices, Facilities: facilities,
			OpenTimes: reception.OpenTimes, Shifts: reception.Shifts,
			StaffSkills: reception.StaffSkills, WorkStaffs: reception.WorkStaffs,
			WaitConfigs: reception.WaitConfigs, Books: books, Queues: queues,
		},
		Menu:         menu,
		FacilityName: seatName,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

// consumeReceptionFacilityPasscode atomically removes a one-time passcode from
// reception.facilities JSONB and registers it as the facility's active code.
func consumeReceptionFacilityPasscode(ctx context.Context, db *sql.DB, receptionID int64, code string) (bool, string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, "", err
	}
	defer tx.Rollback()

	var facilitiesJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT facilities FROM reception WHERE reception_id = $1 FOR UPDATE
	`, receptionID).Scan(&facilitiesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	var facilities []table.FacilityMaster
	if err := json.Unmarshal(facilitiesJSON, &facilities); err != nil {
		return false, "", err
	}

	for i := range facilities {
		for j := range facilities[i].Passcodes {
			if facilities[i].Passcodes[j].Passkey != code {
				continue
			}
			facilityID := facilities[i].FacilityID
			facilityName := facilities[i].FacilityName
			facilities[i].Passcodes = append(facilities[i].Passcodes[:j], facilities[i].Passcodes[j+1:]...)
			updatedJSON, err := json.Marshal(facilities)
			if err != nil {
				return false, "", err
			}
			if _, err = tx.ExecContext(ctx, `
				UPDATE reception SET facilities = $1::jsonb, updated_at = CURRENT_TIMESTAMP
				WHERE reception_id = $2
			`, updatedJSON, receptionID); err != nil {
				return false, "", err
			}
			// current_code is globally unique. A code collision with another
			// facility fails instead of silently moving that code to this facility.
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO reception_facility_code (current_code, reception_id, facility_id, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
				ON CONFLICT (facility_id) DO UPDATE
				SET current_code = EXCLUDED.current_code,
				    reception_id = EXCLUDED.reception_id,
				    updated_at = CURRENT_TIMESTAMP
			`, code, receptionID, facilityID); err != nil {
				return false, "", err
			}
			if err = tx.Commit(); err != nil {
				return false, "", err
			}
			return true, facilityName, nil
		}
	}
	return false, "", nil
}

func loadReceptionBase(ctx context.Context, q receptionQueryer, receptionID int64) (table.ReceptionStruct, error) {
	var reception table.ReceptionStruct
	err := q.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, admin_names, join_names, inquiry_names,
		       order_user_ids, COALESCE(updated_at, CURRENT_TIMESTAMP),
		       COALESCE(reception_title, ''), skills, work_staff_need
		FROM reception WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID, &reception.ChannelID,
		pq.Array(&reception.AdminNames), pq.Array(&reception.JoinNames),
		pq.Array(&reception.InquiryNames),
		pq.Array(&reception.OrderUserIDs), &reception.UpdatedAt,
		&reception.ReceptionTitle, pq.Array(&reception.Skills), &reception.WorkStaffNeed,
	)
	if err != nil {
		return reception, err
	}
	return reception, nil
}

func loadReception(ctx context.Context, q receptionQueryer, receptionID int64) (table.ReceptionStruct, error) {
	var reception table.ReceptionStruct
	var passcodesJSON, asksJSON, askChoicesJSON, askMultiChoicesJSON []byte
	var facilitiesJSON, openTimesJSON, shiftsJSON, staffSkillsJSON []byte
	var workStaffsJSON, waitConfigsJSON, menusJSON, itemDetailsJSON []byte

	err := q.QueryRowContext(ctx, `
		SELECT reception_id, channel_id, admin_names, join_names, inquiry_names,
		       order_user_ids, COALESCE(updated_at, CURRENT_TIMESTAMP),
		       COALESCE(reception_title, ''), skills, work_staff_need,
		       passcodes, asks, ask_choices, ask_multi_choices, facilities,
		       open_times, shifts, staff_skills, work_staffs, wait_configs, menus, item_details
		FROM reception WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID, &reception.ChannelID,
		pq.Array(&reception.AdminNames), pq.Array(&reception.JoinNames),
		pq.Array(&reception.InquiryNames),
		pq.Array(&reception.OrderUserIDs), &reception.UpdatedAt,
		&reception.ReceptionTitle, pq.Array(&reception.Skills), &reception.WorkStaffNeed,
		&passcodesJSON, &asksJSON, &askChoicesJSON, &askMultiChoicesJSON,
		&facilitiesJSON, &openTimesJSON, &shiftsJSON, &staffSkillsJSON,
		&workStaffsJSON, &waitConfigsJSON, &menusJSON, &itemDetailsJSON,
	)
	if err != nil {
		return reception, err
	}

	for _, item := range []struct {
		name string
		raw  []byte
		dst  any
	}{
		{"passcodes", passcodesJSON, &reception.Passcodes},
		{"asks", asksJSON, &reception.Asks},
		{"ask_choices", askChoicesJSON, &reception.AskChoices},
		{"ask_multi_choices", askMultiChoicesJSON, &reception.AskMultiChoices},
		{"facilities", facilitiesJSON, &reception.Facilities},
		{"open_times", openTimesJSON, &reception.OpenTimes},
		{"shifts", shiftsJSON, &reception.Shifts},
		{"staff_skills", staffSkillsJSON, &reception.StaffSkills},
		{"work_staffs", workStaffsJSON, &reception.WorkStaffs},
		{"wait_configs", waitConfigsJSON, &reception.WaitConfigs},
		{"menus", menusJSON, &reception.Menus},
		{"item_details", itemDetailsJSON, &reception.ItemDetails},
	} {
		if len(item.raw) == 0 {
			continue
		}
		if err := json.Unmarshal(item.raw, item.dst); err != nil {
			return reception, err
		}
	}

	return reception, nil
}

func loadReceptionFacilityViews(ctx context.Context, q receptionQueryer, receptionID int64, masters []table.FacilityMaster) ([]table.FacilityView, error) {
	views := make([]table.FacilityView, 0, len(masters))
	for _, master := range masters {
		view := table.FacilityView{
			FacilityID: master.FacilityID, FacilityName: master.FacilityName,
			Capacity: master.Capacity, Passcodes: master.Passcodes, Bookable: master.Bookable,
		}
		err := q.QueryRowContext(ctx, `
			SELECT current_code FROM reception_facility_code
			WHERE reception_id = $1 AND facility_id = $2
		`, receptionID, master.FacilityID).Scan(&view.CurrentCode)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		rows, err := q.QueryContext(ctx, `
			SELECT book_start, book_end
			FROM reception_facility_book_time
			WHERE reception_id = $1 AND facility_id = $2
			ORDER BY sort_order, facility_book_time_id
		`, receptionID, master.FacilityID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var bt table.BookTime
			if err := rows.Scan(&bt.BookStart, &bt.BookEnd); err != nil {
				rows.Close()
				return nil, err
			}
			view.BookTimes = append(view.BookTimes, bt)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
		views = append(views, view)
	}
	return views, nil
}

func loadReceptionBooks(ctx context.Context, q receptionQueryer, receptionID int64) ([]receptionBookResponse, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT book_id, book_start, book_end, answers, COALESCE(menu_id, 0),
		       COALESCE(people, 0), COALESCE(created_at, CURRENT_TIMESTAMP)
		FROM reception_book WHERE reception_id = $1 ORDER BY sort_order, book_id
	`, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := make([]receptionBookResponse, 0)
	for rows.Next() {
		var item receptionBookResponse
		if err := rows.Scan(&item.BookID, &item.BookStart, &item.BookEnd,
			pq.Array(&item.Answers), &item.MenuID, &item.People, &item.CreatedAt); err != nil {
			return nil, err
		}
		books = append(books, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}

func loadReceptionQueues(ctx context.Context, q receptionQueryer, receptionID int64) ([]receptionQueueResponse, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT COALESCE(waiting_guest, 0), COALESCE(queue_name, ''),
		       COALESCE(queued_at, ''), COALESCE(user_id, 0)
		FROM reception_queue WHERE reception_id = $1 ORDER BY sort_order, queue_id
	`, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	queues := make([]receptionQueueResponse, 0)
	for rows.Next() {
		var item receptionQueueResponse
		if err := rows.Scan(&item.WaitingGuest, &item.QueueName, &item.QueuedAt, &item.UserID); err != nil {
			return nil, err
		}
		queues = append(queues, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return queues, nil
}
