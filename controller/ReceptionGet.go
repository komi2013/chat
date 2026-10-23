package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"

	"chat/collection"
	"chat/common"
)

// receptionQueryer is implemented by both *sql.DB and *sql.Tx.
type receptionQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
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

	// Load the reception's channel first. channelID in the request is optional
	// because public reception and booking screens do not need to send it.
	baseReception, err := loadReceptionBase(ctx, db, receptionID)
	if err != nil {
		message := err.Error()
		if err == sql.ErrNoRows {
			message = "Reception not found"
		}
		common.WriteResponseWithSession(w, session, message, http.StatusOK)
		return
	}

	// channelID + aliasName only identify a staff request if they match the
	// logged-in user's row in the alias table and the reception's channel.
	// A failed staff match does not block public access by itself.
	staffAccess := false
	if aliasName != "" && requestedChannelID > 0 && requestedChannelID == baseReception.ChannelID {
		var aliasExists bool
		err = common.DB.AliasDB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM alias
				WHERE user_id = $1
				  AND alias_name = $2
				  AND channel_id = $3
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
		// Public mode: no code is required by design.
	case "2":
		if code != "" {
			validCode, _, err = consumeReceptionFacilityPasscode(ctx, db, receptionID, code)
			if err != nil {
				common.WriteResponseWithSession(w, session, "Passcode update failed: "+err.Error(), http.StatusOK)
				return
			}
		}
	case "3":
		if code != "" {
			err = db.QueryRowContext(ctx, `
				SELECT COALESCE(facility_name, '')
				FROM reception_facility
				WHERE reception_id = $1
				  AND current_code = $2
			ORDER BY sort_order, facility_id
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
		// codeType 1 is intentionally not implemented yet. Non-staff users
		// cannot use this type to bypass code validation.
	default:
		// Unknown code types are not considered valid for non-staff access.
	}

	// codeType == "" deliberately allows public access. For non-public modes,
	// staff may access without a matching customer code, matching the old behavior.
	if !staffAccess && codeType != "" && !validCode {
		common.WriteResponseWithSession(w, session, "コードが一致していません。スタッフでもありません", http.StatusOK)
		return
	}

	// Reload after codeType 2 processing so the consumed passcode and current_code
	// are reflected in the response.
	reception, err := loadReception(ctx, db, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load reception: "+err.Error(), http.StatusOK)
		return
	}

	menu, err := loadReceptionMenu(ctx, db, receptionID)
	if err != nil {
		common.WriteResponseWithSession(w, session, "Failed to load menu: "+err.Error(), http.StatusOK)
		return
	}

	// For non-staff users, hide staff-only fields and facility passcodes.
	if !staffAccess {
		reception.ChannelID = 0
		reception.AdminNames = nil
		reception.Passcodes = nil
		reception.JoinNames = nil
		// These fields can contain staff/customer activity and are staff-only.
		reception.InquiryNames = nil
		reception.Subscriptions = nil
		reception.OrderUserIDs = nil
		for i := range reception.Facilities {
			reception.Facilities[i].Passcodes = nil
		}
	}

	// Nickname is no longer stored in SessionStruct. UserID is an exact, stable
	// account identifier and replaces it as a machine identifier in this response.
	responseData := struct {
		Csrf         string                     `json:"csrf"`
		PushContents []string                   `json:"pushContents"`
		Mail         string                     `json:"mail,omitempty"`
		Telephone    string                     `json:"telephone,omitempty"`
		UserID       int64                      `json:"userID"`
		Reception    collection.ReceptionStruct `json:"reception"`
		Menu         collection.MenuStruct      `json:"menu"`
		FacilityName string                     `json:"facilityName"`
	}{
		Csrf:         session.Csrf,
		PushContents: session.PushContents,
		Mail:         session.Mail,
		Telephone:    session.Telephone,
		UserID:       session.UserID,
		Reception:    reception,
		Menu:         menu,
		FacilityName: seatName,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		return
	}
}

func consumeReceptionFacilityPasscode(ctx context.Context, db *sql.DB, receptionID int64, code string) (bool, string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, "", err
	}
	defer tx.Rollback()

	var passcodeID int64
	var facilityID int64
	var facilityName string
	err = tx.QueryRowContext(ctx, `
		SELECT p.id, f.facility_id, COALESCE(f.facility_name, '')
		FROM reception_facility_passcode AS p
		JOIN reception_facility AS f ON f.facility_id = p.facility_id
		WHERE f.reception_id = $1
		  AND p.passkey = $2
		ORDER BY f.sort_order, p.sort_order, p.id
		LIMIT 1
		FOR UPDATE OF p, f
	`, receptionID, code).Scan(&passcodeID, &facilityID, &facilityName)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM reception_facility_passcode
		WHERE id = $1
	`, passcodeID); err != nil {
		return false, "", err
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE reception_facility
		SET current_code = $1
		WHERE facility_id = $2
		  AND reception_id = $3
	`, code, facilityID, receptionID); err != nil {
		return false, "", err
	}

	if err = tx.Commit(); err != nil {
		return false, "", err
	}
	return true, facilityName, nil
}

func loadReceptionBase(ctx context.Context, q receptionQueryer, receptionID int64) (collection.ReceptionStruct, error) {
	var reception collection.ReceptionStruct
	err := q.QueryRowContext(ctx, `
		SELECT
			reception_id,
			channel_id,
			admin_names,
			join_names,
			inquiry_names,
			subscriptions,
			order_user_ids,
			COALESCE(updated_at, CURRENT_TIMESTAMP),
			COALESCE(reception_title, ''),
			skills,
			work_staff_need
		FROM reception
		WHERE reception_id = $1
	`, receptionID).Scan(
		&reception.ReceptionID,
		&reception.ChannelID,
		pq.Array(&reception.AdminNames),
		pq.Array(&reception.JoinNames),
		pq.Array(&reception.InquiryNames),
		pq.Array(&reception.Subscriptions),
		pq.Array(&reception.OrderUserIDs),
		&reception.UpdatedAt,
		&reception.ReceptionTitle,
		pq.Array(&reception.Skills),
		&reception.WorkStaffNeed,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return reception, sql.ErrNoRows
		}
		return reception, err
	}
	return reception, nil
}

func loadReception(ctx context.Context, q receptionQueryer, receptionID int64) (collection.ReceptionStruct, error) {
	reception, err := loadReceptionBase(ctx, q, receptionID)
	if err != nil {
		return reception, err
	}

	// Reception passcodes.
	rows, err := q.QueryContext(ctx, `
		SELECT passkey, COALESCE(pass_start, ''), COALESCE(pass_end, '')
		FROM reception_passcode
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var p collection.Passcode
		if err := rows.Scan(&p.Passkey, &p.PassStart, &p.PassEnd); err != nil {
			rows.Close()
			return reception, err
		}
		reception.Passcodes = append(reception.Passcodes, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Open times.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(limit_start, ''), COALESCE(limit_end, '')
		FROM reception_open_time
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.OpenTime
		if err := rows.Scan(&v.LimitStart, &v.LimitEnd); err != nil {
			rows.Close()
			return reception, err
		}
		reception.OpenTimes = append(reception.OpenTimes, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Questions.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(question, ''), sequence
		FROM reception_ask
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.Ask
		if err := rows.Scan(&v.Question, &v.Sequence); err != nil {
			rows.Close()
			return reception, err
		}
		reception.Asks = append(reception.Asks, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Single-choice questions.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(question, ''), choices, sequence
		FROM reception_ask_choice
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.AskChoice
		if err := rows.Scan(&v.Question, pq.Array(&v.Choices), &v.Sequence); err != nil {
			rows.Close()
			return reception, err
		}
		reception.AskChoices = append(reception.AskChoices, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Multiple-choice questions.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(question, ''), choices, sequence
		FROM reception_ask_multi_choice
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.AskMultiChoice
		if err := rows.Scan(&v.Question, pq.Array(&v.Choices), &v.Sequence); err != nil {
			rows.Close()
			return reception, err
		}
		reception.AskMultiChoices = append(reception.AskMultiChoices, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Facilities. Keep their database IDs locally to fetch nested records.
	var facilityIDs []int64
	rows, err = q.QueryContext(ctx, `
		SELECT facility_id, COALESCE(facility_name, ''), COALESCE(capacity, 0),
		       COALESCE(current_code, ''), COALESCE(bookable, FALSE)
		FROM reception_facility
		WHERE reception_id = $1
		ORDER BY sort_order, facility_id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var id int64
		var v collection.Facility
		if err := rows.Scan(&id, &v.FacilityName, &v.Capacity, &v.CurrentCode, &v.Bookable); err != nil {
			rows.Close()
			return reception, err
		}
		facilityIDs = append(facilityIDs, id)
		reception.Facilities = append(reception.Facilities, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	for i, facilityID := range facilityIDs {
		rows, err = q.QueryContext(ctx, `
			SELECT passkey, COALESCE(pass_start, ''), COALESCE(pass_end, '')
			FROM reception_facility_passcode
			WHERE facility_id = $1
			ORDER BY sort_order, id
		`, facilityID)
		if err != nil {
			return reception, err
		}
		for rows.Next() {
			var p collection.Passcode
			if err := rows.Scan(&p.Passkey, &p.PassStart, &p.PassEnd); err != nil {
				rows.Close()
				return reception, err
			}
			reception.Facilities[i].Passcodes = append(reception.Facilities[i].Passcodes, p)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return reception, err
		}
		rows.Close()

		rows, err = q.QueryContext(ctx, `
			SELECT COALESCE(book_start, ''), COALESCE(book_end, '')
			FROM reception_facility_book_time
			WHERE facility_id = $1
			ORDER BY sort_order, id
		`, facilityID)
		if err != nil {
			return reception, err
		}
		for rows.Next() {
			var bt collection.BookTime
			if err := rows.Scan(&bt.BookStart, &bt.BookEnd); err != nil {
				rows.Close()
				return reception, err
			}
			reception.Facilities[i].BookTimes = append(reception.Facilities[i].BookTimes, bt)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return reception, err
		}
		rows.Close()
	}

	// Shifts.
	rows, err = q.QueryContext(ctx, `
		SELECT shift_id, alias_names, COALESCE(shift_start, ''), COALESCE(shift_end, ''), open,
		       COALESCE(skill, ''), fix
		FROM reception_shift
		WHERE reception_id = $1
		ORDER BY sort_order, shift_id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.Shift
		if err := rows.Scan(&v.ShiftID, pq.Array(&v.AliasNames), &v.ShiftStart, &v.ShiftEnd, &v.Open, &v.Skill, &v.Fix); err != nil {
			rows.Close()
			return reception, err
		}
		reception.Shifts = append(reception.Shifts, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Staff skills.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(alias_name, ''), skills
		FROM reception_staff_skill
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.StaffSkill
		if err := rows.Scan(&v.AliasName, pq.Array(&v.Skills)); err != nil {
			rows.Close()
			return reception, err
		}
		reception.StaffSkills = append(reception.StaffSkills, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Work staff.
	rows, err = q.QueryContext(ctx, `
		SELECT shift_id, COALESCE(alias_name, ''), COALESCE(work_start, ''), COALESCE(work_end, ''), seq
		FROM reception_work_staff
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.WorkStaff
		if err := rows.Scan(&v.ShiftID, &v.AliasName, &v.WorkStart, &v.WorkEnd, &v.Seq); err != nil {
			rows.Close()
			return reception, err
		}
		reception.WorkStaffs = append(reception.WorkStaffs, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Books.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(book_start, ''), COALESCE(book_end, ''), answers,
		       COALESCE(menu_id, 0), COALESCE(people, 0),
		       COALESCE(created_at, CURRENT_TIMESTAMP), COALESCE(nickname, '')
		FROM reception_book
		WHERE reception_id = $1
		ORDER BY sort_order, book_id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.Book
		if err := rows.Scan(&v.BookStart, &v.BookEnd, pq.Array(&v.Answers), &v.MenuID, &v.People, &v.CreatedAt, &v.Nickname); err != nil {
			rows.Close()
			return reception, err
		}
		reception.Books = append(reception.Books, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Queue entries.
	rows, err = q.QueryContext(ctx, `
		SELECT COALESCE(waiting_guest, 0), COALESCE(queue_name, ''),
		       COALESCE(queued_at, ''), COALESCE(user_id, 0)
		FROM reception_queue
		WHERE reception_id = $1
		ORDER BY sort_order, queue_id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var v collection.Queue
		if err := rows.Scan(&v.WaitingGuest, &v.QueueName, &v.QueuedAt, &v.UserID); err != nil {
			rows.Close()
			return reception, err
		}
		reception.Queues = append(reception.Queues, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	// Wait configs store INTEGER[] while the public struct uses [2]int.
	rows, err = q.QueryContext(ctx, `
		SELECT guest_range, COALESCE(wait_ratio, 0)
		FROM reception_wait_config
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return reception, err
	}
	for rows.Next() {
		var values []int
		var v collection.WaitConfig
		if err := rows.Scan(pq.Array(&values), &v.WaitRatio); err != nil {
			rows.Close()
			return reception, err
		}
		if len(values) > 0 {
			v.GuestRange[0] = values[0]
		}
		if len(values) > 1 {
			v.GuestRange[1] = values[1]
		}
		reception.WaitConfigs = append(reception.WaitConfigs, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return reception, err
	}
	rows.Close()

	return reception, nil
}

func loadReceptionMenu(ctx context.Context, q receptionQueryer, receptionID int64) (collection.MenuStruct, error) {
	menu := collection.MenuStruct{ReceptionID: receptionID}
	err := q.QueryRowContext(ctx, `
		SELECT COALESCE(updated_at, CURRENT_TIMESTAMP)
		FROM reception_menu
		WHERE reception_id = $1
	`, receptionID).Scan(&menu.UpdatedAt)
	if err == sql.ErrNoRows {
		return menu, nil
	}
	if err != nil {
		return menu, err
	}

	type menuItemRecord struct {
		rowID int64
		item  collection.Menu
	}
	var menuItems []menuItemRecord
	rows, err := q.QueryContext(ctx, `
		SELECT id, menu_id, COALESCE(menu_name, ''), COALESCE(price, 0),
		       COALESCE(prepaid_price, 0), COALESCE(need_skill, ''),
		       COALESCE(need_facility, ''), COALESCE(specify_name_flag, FALSE),
		       COALESCE(spend_minute, 0), COALESCE(for_book_type, 0)
		FROM reception_menu_item
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return menu, err
	}
	for rows.Next() {
		var rec menuItemRecord
		if err := rows.Scan(
			&rec.rowID,
			&rec.item.MenuID,
			&rec.item.MenuName,
			&rec.item.Price,
			&rec.item.PrepaidPrice,
			&rec.item.NeedSkill,
			&rec.item.NeedFacility,
			&rec.item.SpecifyNameFlag,
			&rec.item.SpendMinute,
			&rec.item.ForBookType,
		); err != nil {
			rows.Close()
			return menu, err
		}
		menuItems = append(menuItems, rec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return menu, err
	}
	rows.Close()

	for _, rec := range menuItems {
		var err error
		rec.item.Items, err = loadInt64Values(ctx, q, `
			SELECT item_id FROM reception_menu_item_content
			WHERE menu_item_id = $1 ORDER BY sort_order, id
		`, rec.rowID)
		if err != nil {
			return menu, err
		}

		rows, err = q.QueryContext(ctx, `
			SELECT item_id, price FROM reception_menu_paid_option
			WHERE menu_item_id = $1 ORDER BY sort_order, id
		`, rec.rowID)
		if err != nil {
			return menu, err
		}
		for rows.Next() {
			var option collection.ItemOption
			if err := rows.Scan(&option.ItemID, &option.Price); err != nil {
				rows.Close()
				return menu, err
			}
			rec.item.PaidOptions = append(rec.item.PaidOptions, option)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return menu, err
		}
		rows.Close()

		groupRows, err := q.QueryContext(ctx, `
			SELECT id FROM reception_menu_free_option_group
			WHERE menu_item_id = $1 ORDER BY sort_order, id
		`, rec.rowID)
		if err != nil {
			return menu, err
		}
		var groupIDs []int64
		for groupRows.Next() {
			var groupID int64
			if err := groupRows.Scan(&groupID); err != nil {
				groupRows.Close()
				return menu, err
			}
			groupIDs = append(groupIDs, groupID)
		}
		if err := groupRows.Err(); err != nil {
			groupRows.Close()
			return menu, err
		}
		groupRows.Close()

		for _, groupID := range groupIDs {
			values, err := loadInt64Values(ctx, q, `
				SELECT item_id FROM reception_menu_free_option_item
				WHERE free_option_group_id = $1 ORDER BY sort_order, id
			`, groupID)
			if err != nil {
				return menu, err
			}
			rec.item.FreeOptions = append(rec.item.FreeOptions, values)
		}

		rec.item.FreeMultiOptions, err = loadInt64Values(ctx, q, `
			SELECT item_id FROM reception_menu_free_multi_option
			WHERE menu_item_id = $1 ORDER BY sort_order, id
		`, rec.rowID)
		if err != nil {
			return menu, err
		}
		menu.Menus = append(menu.Menus, rec.item)
	}

	// Item details, including nested choices.
	type detailRecord struct {
		rowID  int64
		detail collection.ItemDetail
	}
	var details []detailRecord
	rows, err = q.QueryContext(ctx, `
		SELECT id, item_id, COALESCE(item_name, ''), COALESCE(img_path, '')
		FROM reception_menu_item_detail
		WHERE reception_id = $1
		ORDER BY sort_order, id
	`, receptionID)
	if err != nil {
		return menu, err
	}
	for rows.Next() {
		var rec detailRecord
		if err := rows.Scan(&rec.rowID, &rec.detail.ItemID, &rec.detail.ItemName, &rec.detail.ImgPath); err != nil {
			rows.Close()
			return menu, err
		}
		details = append(details, rec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return menu, err
	}
	rows.Close()

	for _, rec := range details {
		groupRows, err := q.QueryContext(ctx, `
			SELECT id FROM reception_menu_item_detail_choice_group
			WHERE item_detail_id = $1 ORDER BY sort_order, id
		`, rec.rowID)
		if err != nil {
			return menu, err
		}
		var groupIDs []int64
		for groupRows.Next() {
			var groupID int64
			if err := groupRows.Scan(&groupID); err != nil {
				groupRows.Close()
				return menu, err
			}
			groupIDs = append(groupIDs, groupID)
		}
		if err := groupRows.Err(); err != nil {
			groupRows.Close()
			return menu, err
		}
		groupRows.Close()

		for _, groupID := range groupIDs {
			choices := make([]string, 0)
			rows, err = q.QueryContext(ctx, `
				SELECT choice_text FROM reception_menu_item_detail_choice
				WHERE choice_group_id = $1 ORDER BY sort_order, id
			`, groupID)
			if err != nil {
				return menu, err
			}
			for rows.Next() {
				var choice string
				if err := rows.Scan(&choice); err != nil {
					rows.Close()
					return menu, err
				}
				choices = append(choices, choice)
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return menu, err
			}
			rows.Close()
			rec.detail.Choices = append(rec.detail.Choices, choices)
		}
		menu.ItemDetails = append(menu.ItemDetails, rec.detail)
	}

	return menu, nil
}

func loadInt64Values(ctx context.Context, q receptionQueryer, query string, parentID int64) ([]int64, error) {
	rows, err := q.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]int64, 0)
	for rows.Next() {
		var value int64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
